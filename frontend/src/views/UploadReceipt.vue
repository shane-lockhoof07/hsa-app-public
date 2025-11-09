<template>
  <div>
    <!-- Summary Header -->
    <v-row class="mb-2">
      <v-col cols="6" md="6">
        <v-card color="success" variant="tonal" class="rounded-xl">
          <v-card-text>
            <div class="text-h6 text-success-darken-2">Available</div>
            <div class="text-h4 font-weight-bold">
              ${{ amountAvailable.toFixed(2) }}
            </div>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="6" md="6">
        <v-card color="warning" variant="tonal" class="rounded-xl">
          <v-card-text>
            <div class="text-h6 text-warning-darken-2">Used</div>
            <div class="text-h4 font-weight-bold">
              ${{ amountUsed.toFixed(2) }}
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Upload Card -->
    <v-card variant="outlined" max-width="600" class="mx-auto rounded-xl">
      <v-card-title class="text-center text-h6 pa-4">
        Upload a Receipt
      </v-card-title>

      <v-card-text class="pa-8">
        <div
          @click="triggerFileInput"
          @dragover.prevent="isDragging = true"
          @dragleave.prevent="isDragging = false"
          @drop.prevent="handleDrop"
          :class="['upload-area', { 'upload-area--dragging': isDragging }]"
        >
          <v-icon v-if="!previewUrl && !isPDF" size="80" color="grey-lighten-1">
            mdi-cloud-upload-outline
          </v-icon>

          <!-- PDF Preview -->
          <div v-if="isPDF" class="pdf-preview">
            <v-icon size="80" color="red-darken-2"> mdi-file-pdf-box </v-icon>
            <div class="text-h6 mt-2">{{ file?.name }}</div>
          </div>

          <!-- Image Preview -->
          <v-img
            v-if="previewUrl && !isPDF"
            :src="previewUrl"
            max-height="300"
            contain
            class="mb-4"
          ></v-img>

          <div
            v-if="!previewUrl && !isPDF"
            class="text-h6 text-grey-lighten-1 mt-4"
          >
            Click to upload or drag and drop
          </div>
          <div v-if="!previewUrl && !isPDF" class="text-body-2 text-grey mt-2">
            JPG, PNG, HEIC, or PDF
          </div>
        </div>

        <input
          ref="fileInput"
          type="file"
          accept="image/*,application/pdf"
          @change="handleFileSelect"
          style="display: none"
        />

        <v-alert v-if="message" :type="messageType" class="mt-4">
          {{ message }}
        </v-alert>

        <div v-if="file" class="mt-4">
          <v-btn
            color="primary"
            block
            size="large"
            :loading="uploading"
            @click="upload"
          >
            <v-icon left>mdi-upload</v-icon>
            Upload Receipt
          </v-btn>

          <v-btn variant="text" block class="mt-2" @click="clearFile">
            Cancel
          </v-btn>
        </div>
      </v-card-text>
    </v-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import api from "../services/api";

const file = ref(null);
const previewUrl = ref(null);
const isPDF = ref(false);
const uploading = ref(false);
const message = ref("");
const messageType = ref("info");
const isDragging = ref(false);
const fileInput = ref(null);
const amountAvailable = ref(0);
const amountUsed = ref(0);

const triggerFileInput = () => {
  fileInput.value.click();
};

const handleFileSelect = (event) => {
  const selectedFile = event.target.files[0];
  if (selectedFile) {
    processFile(selectedFile);
  }
};

const handleDrop = (event) => {
  isDragging.value = false;
  const droppedFile = event.dataTransfer.files[0];
  if (droppedFile) {
    processFile(droppedFile);
  }
};

const processFile = (selectedFile) => {
  file.value = selectedFile;

  // Check if PDF
  if (selectedFile.type === "application/pdf") {
    isPDF.value = true;
    previewUrl.value = null;
  } else {
    isPDF.value = false;
    const reader = new FileReader();
    reader.onload = (e) => {
      previewUrl.value = e.target.result;
    };
    reader.readAsDataURL(selectedFile);
  }
};

const clearFile = () => {
  file.value = null;
  previewUrl.value = null;
  isPDF.value = false;
  message.value = "";
  if (fileInput.value) {
    fileInput.value.value = "";
  }
};

const loadSummary = async () => {
  try {
    const receipts = await api.getReceipts();
    amountAvailable.value = receipts
      .filter((r) => !r.used && r.hsa_qualified)
      .reduce((sum, r) => sum + r.total_amount, 0);
    amountUsed.value = receipts
      .filter((r) => r.used)
      .reduce((sum, r) => sum + r.total_amount, 0);
  } catch (error) {
    console.error("Failed to load summary:", error);
  }
};

const upload = async () => {
  if (!file.value) return;

  uploading.value = true;
  message.value = "";

  try {
    const result = await api.uploadReceipt(file.value);
    messageType.value = "success";
    message.value = `Receipt uploaded successfully! Vendor: ${result.vendor}, Amount: $${result.amount}`;

    // Reload summary
    await loadSummary();

    // Reset form after a delay
    setTimeout(() => {
      clearFile();
    }, 2000);
  } catch (error) {
    messageType.value = "error";
    // Show the detailed error message from the API
    message.value =
      error.message ||
      `Upload failed: ${error.response?.data?.message || "Unknown error"}`;
  } finally {
    uploading.value = false;
  }
};

onMounted(() => {
  loadSummary();
});
</script>

<style scoped>
.upload-area {
  border: 2px dashed #e0e0e0;
  border-radius: 8px;
  padding: 48px 24px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
  min-height: 280px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.upload-area:hover {
  border-color: #9e9e9e;
  background-color: #fafafa;
}

.upload-area--dragging {
  border-color: #1976d2;
  background-color: #e3f2fd;
}

.pdf-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
</style>
