<template>
  <v-card max-width="800" class="mx-auto rounded-lg">
    <v-card-title class="d-flex align-center">
      <v-row>
        <v-col cols="12" sm="6">
          <span class="text-h5">HSA Deduction Calculator</span>
        </v-col>
        <v-col cols="12" sm="6" class="d-flex align-center justify-end">
          <v-chip color="success" variant="tonal">
            <v-icon left size="small">mdi-check-circle</v-icon>
            Available: ${{ (amountAvailable || 0).toFixed(2) }}
          </v-chip>
        </v-col>
      </v-row>
    </v-card-title>
    <v-card-text>
      <v-text-field
        v-model.number="targetAmount"
        label="Target Deduction Amount"
        type="number"
        prefix="$"
        variant="outlined"
      ></v-text-field>

      <v-btn
        color="primary"
        :loading="calculating"
        :disabled="!targetAmount || approved"
        @click="calculate"
        block
        class="mb-4"
      >
        Calculate Optimal Receipts
      </v-btn>

      <v-alert v-if="error" type="error" class="mb-4">{{ error }}</v-alert>

      <div v-if="selectedReceipts.length > 0 && !approved">
        <v-alert type="success" class="mb-4">
          <strong>Total Selected: ${{ totalSelected.toFixed(2) }}</strong>
          ({{ selectedReceipts.length }} receipt{{
            selectedReceipts.length > 1 ? "s" : ""
          }})
        </v-alert>

        <v-list class="mb-4">
          <v-list-item
            v-for="receipt in selectedReceipts"
            :key="receipt.id"
            class="mb-2"
          >
            <v-card>
              <v-card-text>
                <v-row>
                  <v-col cols="4">
                    <strong>{{ receipt.vendor }}</strong>
                  </v-col>
                  <v-col cols="4">
                    {{ formatDateUTC(receipt.date) }}
                  </v-col>
                  <v-col cols="4" class="text-right">
                    <v-chip color="success" size="small">
                      ${{ receipt.total_amount.toFixed(2) }}
                    </v-chip>
                  </v-col>
                </v-row>
              </v-card-text>
            </v-card>
          </v-list-item>
        </v-list>

        <!-- Approve Button -->
        <v-btn
          color="success"
          :loading="approving"
          @click="approveDeduction"
          block
          size="large"
          class="mb-4"
        >
          <v-icon left>mdi-check-circle</v-icon>
          Approve & Mark as Used
        </v-btn>
      </div>

      <!-- Success message after approval -->
      <v-alert v-if="approved" type="success" class="mb-4">
        <div class="text-h6">✓ Deduction Approved!</div>
        <div class="mt-2">
          {{ selectedReceipts.length }} receipt{{
            selectedReceipts.length > 1 ? "s" : ""
          }}
          totaling ${{ totalSelected.toFixed(2) }} have been marked as used.
        </div>
      </v-alert>
    </v-card-text>

    <!-- Button moved outside the card at bottom -->
    <v-card-actions v-if="approved" class="pa-4">
      <v-btn color="primary" variant="flat" @click="reset" block size="large">
        <v-icon left>mdi-calculator</v-icon>
        Calculate Another Deduction
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import api from "../services/api";

const targetAmount = ref(null);
const selectedReceipts = ref([]);
const calculating = ref(false);
const approving = ref(false);
const approved = ref(false);
const error = ref("");
const amountAvailable = ref(0);

const totalSelected = computed(() => {
  return selectedReceipts.value.reduce(
    (sum, receipt) => sum + receipt.total_amount,
    0
  );
});

// Load available balance from the database
const loadAvailableBalance = async () => {
  try {
    const allReceipts = await api.getReceipts();
    amountAvailable.value = allReceipts
      .filter((r) => !r.used && r.hsa_qualified)
      .reduce((sum, r) => sum + r.total_amount, 0);
  } catch (err) {
    console.error("Failed to load available balance:", err);
  }
};

// Fix for timezone offset issue
const formatDateUTC = (dateString) => {
  const date = new Date(dateString);
  const userTimezoneOffset = date.getTimezoneOffset() * 60000;
  const correctedDate = new Date(date.getTime() + userTimezoneOffset);
  return correctedDate.toLocaleDateString();
};

const calculate = async () => {
  if (!targetAmount.value) return;

  calculating.value = true;
  error.value = "";
  approved.value = false;

  try {
    const receipts = await api.calculateDeduction(targetAmount.value);
    selectedReceipts.value = receipts;

    if (receipts.length === 0) {
      error.value = "No combination of receipts matches the target amount.";
    }
  } catch (err) {
    error.value = `Failed to calculate: ${err.message}`;
    selectedReceipts.value = [];
  } finally {
    calculating.value = false;
  }
};

const approveDeduction = async () => {
  approving.value = true;
  error.value = "";

  try {
    const receiptIds = selectedReceipts.value.map((r) => r.id);
    console.log("Approving receipt IDs:", receiptIds);
    
    await api.approveDeduction(receiptIds);
    
    console.log("Approval complete, reloading receipts...");
    approved.value = true;

    // Update available amount after approval
    await loadAvailableBalance();
    
    console.log("Balance updated");
  } catch (err) {
    console.error("Approval error:", err);
    error.value = `Failed to approve deduction: ${err.message}`;
  } finally {
    approving.value = false;
  }
};

const reset = () => {
  targetAmount.value = null;
  selectedReceipts.value = [];
  approved.value = false;
  error.value = "";
};

// Load available balance when component mounts
onMounted(() => {
  loadAvailableBalance();
});
</script>