<template>
  <div>
    <v-card class="mb-4">
      <v-card-title>Receipt Summary</v-card-title>
      <v-card-text>
        <v-row>
          <v-col cols="12" md="4">
            <v-card color="success" variant="tonal">
              <v-card-text>
                <div class="text-h6">Available</div>
                <div class="text-h4">${{ amountAvailable.toFixed(2) }}</div>
              </v-card-text>
            </v-card>
          </v-col>
          <v-col cols="12" md="4">
            <v-card color="warning" variant="tonal">
              <v-card-text>
                <div class="text-h6">Used</div>
                <div class="text-h4">${{ amountUsed.toFixed(2) }}</div>
              </v-card-text>
            </v-card>
          </v-col>
          <v-col cols="12" md="4">
            <v-card color="primary" variant="tonal">
              <v-card-text>
                <div class="text-h6">Total Receipts</div>
                <div class="text-h4">{{ receipts.length }}</div>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <v-card>
      <v-card-title>
        <v-icon class="mr-2">mdi-receipt-text-outline</v-icon>
        All Receipts
      </v-card-title>

      <v-card-text>
        <v-alert
          v-if="error"
          type="error"
          dismissible
          @click:close="error = null"
        >
          {{ error }}
        </v-alert>

        <v-data-table
          :headers="headers"
          :items="receipts"
          :loading="loading"
          item-value="id"
          @click:row="viewReceipt"
          hover
          class="elevation-1"
        >
          <template v-slot:item.date="{ item }">
            {{ new Date(item.date).toLocaleDateString() }}
          </template>

          <template v-slot:item.total_amount="{ item }">
            ${{ item.total_amount.toFixed(2) }}
          </template>

          <template v-slot:item.hsa_status="{ item }">
            <v-chip
              :color="
                item.hsa_status === 'Yes'
                  ? 'success'
                  : item.hsa_status === 'Partially'
                  ? 'warning'
                  : 'error'
              "
              size="small"
            >
              {{ item.hsa_status }}
            </v-chip>
          </template>

          <template v-slot:item.use_reason="{ item }">
            {{ item.use_reason || "-" }}
          </template>

          <template v-slot:item.used="{ item }">
            <v-chip :color="item.used ? 'warning' : 'success'" size="small">
              {{ item.used ? "Used" : "Available" }}
            </v-chip>
          </template>
        </v-data-table>
      </v-card-text>
    </v-card>

    <!-- Edit Receipt Dialog -->
    <v-dialog v-model="dialog" max-width="900px" scrollable>
      <v-card>
        <v-card-title>
          <span class="text-h5">Receipt Details</span>
          <v-spacer></v-spacer>
          <v-btn icon @click="dialog = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>

        <v-card-text>
          <v-row>
            <v-col cols="12" md="6">
              <!-- Receipt Image Preview -->
              <div class="receipt-preview">
                <!-- For PDF receipts -->
                <div v-if="isPDFReceipt(selectedReceipt)" class="pdf-viewer">
                  <v-icon size="80" color="red-darken-2">
                    mdi-file-pdf-box
                  </v-icon>
                  <div class="text-h6 mt-4">PDF Receipt</div>
                  <v-btn
                    :href="getReceiptImageUrl(selectedReceipt)"
                    target="_blank"
                    color="primary"
                    class="mt-4"
                  >
                    <v-icon left>mdi-file-download</v-icon>
                    Open PDF
                  </v-btn>
                </div>

                <!-- For Image receipts (JPG, PNG, HEIC) -->
                <div v-else class="image-viewer">
                  <v-img
                    v-if="selectedReceipt"
                    :src="getReceiptImageUrl(selectedReceipt)"
                    :alt="`${selectedReceipt.vendor} receipt`"
                    max-height="500"
                    contain
                    class="rounded"
                  >
                    <template v-slot:placeholder>
                      <v-row
                        class="fill-height ma-0"
                        align="center"
                        justify="center"
                      >
                        <v-progress-circular
                          indeterminate
                          color="primary"
                        ></v-progress-circular>
                      </v-row>
                    </template>
                    <template v-slot:error>
                      <v-row
                        class="fill-height ma-0"
                        align="center"
                        justify="center"
                      >
                        <v-icon size="48" color="error"
                          >mdi-alert-circle</v-icon
                        >
                        <div class="text-body-1 mt-2">Failed to load image</div>
                      </v-row>
                    </template>
                  </v-img>

                  <!-- Button to open in new tab -->
                  <v-btn
                    v-if="selectedReceipt"
                    :href="getReceiptImageUrl(selectedReceipt)"
                    target="_blank"
                    variant="text"
                    color="primary"
                    class="mt-2"
                  >
                    <v-icon left>mdi-open-in-new</v-icon>
                    Open Full Size
                  </v-btn>
                </div>
              </div>
            </v-col>

            <v-col cols="12" md="6">
              <v-form>
                <v-text-field
                  v-model="editedReceipt.vendor"
                  label="Vendor"
                  required
                ></v-text-field>

                <v-text-field
                  v-model.number="editedReceipt.total_amount"
                  label="Amount"
                  type="number"
                  step="0.01"
                  prefix="$"
                  required
                ></v-text-field>

                <v-text-field
                  v-model="editedReceipt.date"
                  label="Date"
                  type="date"
                  required
                ></v-text-field>

                <v-select
                  v-model="editedReceipt.hsa_status"
                  :items="['Yes', 'No', 'Partially']"
                  label="HSA Qualified"
                  required
                >
                  <template v-slot:item="{ props, item }">
                    <v-list-item v-bind="props">
                      <template v-slot:prepend>
                        <v-icon
                          :color="
                            item.value === 'Yes'
                              ? 'success'
                              : item.value === 'Partially'
                              ? 'warning'
                              : 'error'
                          "
                        >
                          {{
                            item.value === "Yes"
                              ? "mdi-check-circle"
                              : item.value === "Partially"
                              ? "mdi-alert-circle"
                              : "mdi-close-circle"
                          }}
                        </v-icon>
                      </template>
                    </v-list-item>
                  </template>
                </v-select>

                <!-- Common HSA Expenses Reference -->
                <v-expansion-panels class="mt-3 mb-3">
                  <v-expansion-panel>
                    <v-expansion-panel-title>
                      <v-icon class="mr-2">mdi-information-outline</v-icon>
                      Common HSA-Qualified Expenses
                    </v-expansion-panel-title>
                    <v-expansion-panel-text>
                      <div class="text-subtitle-2 mb-2 text-success-darken-2">
                        <v-icon size="small" color="success"
                          >mdi-check-circle</v-icon
                        >
                        Typically HSA-Qualified:
                      </div>
                      <v-chip-group column>
                        <v-chip
                          v-for="expense in commonHSAExpenses"
                          :key="expense"
                          size="small"
                          color="success"
                          variant="outlined"
                        >
                          {{ expense }}
                        </v-chip>
                      </v-chip-group>
                      <v-divider class="my-3"></v-divider>
                      <div class="text-subtitle-2 mb-2 text-error-darken-2">
                        <v-icon size="small" color="error"
                          >mdi-close-circle</v-icon
                        >
                        Not Qualified:
                      </div>
                      <div class="text-caption text-grey mb-3">
                        Regular food & beverages, household items, vitamins
                        (unless prescribed), cosmetics, toiletries (shampoo,
                        deodorant, toothpaste), general wellness items
                      </div>

                      <v-divider class="my-3"></v-divider>
                      <div class="text-subtitle-2 mb-2 text-primary">
                        <v-icon size="small" color="primary"
                          >mdi-calculator</v-icon
                        >
                        How to Calculate Proportional Tax:
                      </div>
                      <div class="text-caption mb-2">
                        For "Partially" qualified receipts, calculate the
                        HSA-qualified amount including proportional tax:
                      </div>
                      <v-card
                        variant="tonal"
                        color="blue-grey-lighten-5"
                        class="pa-3"
                      >
                        <div class="text-caption font-weight-bold mb-2">
                          Example:
                        </div>
                        <div class="text-caption mb-1">
                          • Band-aids:
                          <span class="font-weight-bold">$10.00</span> (HSA ✓)
                        </div>
                        <div class="text-caption mb-1">
                          • Shampoo:
                          <span class="font-weight-bold">$8.00</span> (Not HSA
                          ✗)
                        </div>
                        <div class="text-caption mb-1">
                          • Subtotal:
                          <span class="font-weight-bold">$18.00</span>
                        </div>
                        <div class="text-caption mb-2">
                          • Tax (8%):
                          <span class="font-weight-bold">$1.44</span>
                        </div>
                        <v-divider class="my-2"></v-divider>
                        <div class="text-caption mb-1">
                          <strong>Step 1:</strong> HSA Items = $10.00
                        </div>
                        <div class="text-caption mb-1">
                          <strong>Step 2:</strong> Proportional Tax = ($10.00 /
                          $18.00) × $1.44 =
                          <span class="font-weight-bold">$0.80</span>
                        </div>
                        <div class="text-caption">
                          <strong>Step 3:</strong> HSA Amount = $10.00 + $0.80 =
                          <span class="font-weight-bold text-success"
                            >$10.80</span
                          >
                        </div>
                      </v-card>

                      <div class="text-caption mt-2 text-grey-darken-1">
                        <v-icon size="x-small">mdi-lightbulb-outline</v-icon>
                        Tip: Only include tax that applies to HSA-qualified
                        items
                      </div>
                    </v-expansion-panel-text>
                  </v-expansion-panel>
                </v-expansion-panels>

                <v-checkbox
                  v-model="editedReceipt.used"
                  label="Mark as Used"
                  color="warning"
                ></v-checkbox>

                <v-text-field
                  v-if="editedReceipt.used"
                  v-model="editedReceipt.use_reason"
                  label="Use Reason (Optional)"
                  placeholder="e.g., Deduction for Q4 2024"
                ></v-text-field>
              </v-form>
            </v-col>
          </v-row>
        </v-card-text>

        <v-card-actions>
          <v-btn color="error" @click="deleteReceipt">
            <v-icon left>mdi-delete</v-icon>
            Delete
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn @click="dialog = false">Cancel</v-btn>
          <v-btn color="primary" @click="saveReceipt" :loading="saving">
            Save
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="text-h5">Confirm Delete</v-card-title>
        <v-card-text>
          Are you sure you want to delete this receipt? This action cannot be
          undone.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" @click="confirmDelete" :loading="deleting">
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import api from "../services/api";

const receipts = ref([]);
const loading = ref(false);
const error = ref(null);
const dialog = ref(false);
const deleteDialog = ref(false);
const selectedReceipt = ref(null);
const editedReceipt = ref({});
const saving = ref(false);
const deleting = ref(false);
const amountAvailable = ref(0);
const amountUsed = ref(0);

// Common HSA-Qualified Expenses for reference
const commonHSAExpenses = [
  "Prescription Medications",
  "Doctor Visits & Copays",
  "Dental Care",
  "Vision Care & Glasses",
  "Contact Lenses & Solution",
  "First Aid Supplies",
  "Bandages & Gauze",
  "Thermometers",
  "Blood Pressure Monitors",
  "Glucose Monitors & Strips",
  "Diabetic Supplies",
  "Insulin",
  "Hearing Aids",
  "Crutches & Braces",
  "Heating Pads",
  "Cold Packs",
  "Pregnancy Tests",
  "Birth Control (prescribed)",
  "Acne Treatment (prescribed)",
  "Allergy Medications",
  "Pain Relievers (OTC)",
  "Sunscreen (SPF 15+)",
  "Medical Masks",
  "Hand Sanitizer",
];

const headers = [
  { title: "Date", key: "date", sortable: true },
  { title: "Vendor", key: "vendor", sortable: true },
  { title: "Amount", key: "total_amount", sortable: true },
  { title: "HSA Status", key: "hsa_status", sortable: true },
  { title: "Status", key: "used", sortable: true },
  { title: "Use Reason", key: "use_reason", sortable: false },
];

const loadReceipts = async () => {
  loading.value = true;
  error.value = null;

  try {
    receipts.value = await api.getReceipts();

    amountAvailable.value = receipts.value
      .filter(
        (r) =>
          !r.used && (r.hsa_status === "Yes" || r.hsa_status === "Partially")
      )
      .reduce((sum, r) => sum + r.total_amount, 0);
    amountUsed.value = receipts.value
      .filter((r) => r.used)
      .reduce((sum, r) => sum + r.total_amount, 0);
  } catch (err) {
    error.value = `Failed to load receipts: ${err.message}`;
  } finally {
    loading.value = false;
  }
};

const viewReceipt = (event, { item }) => {
  selectedReceipt.value = item;
  editedReceipt.value = {
    ...item,
    date: new Date(item.date).toISOString().split("T")[0],
    use_reason: item.use_reason || "",
    hsa_status: item.hsa_status || (item.hsa_qualified ? "Yes" : "No"),
  };

  // Debug: Log the image URL
  const imageUrl = getReceiptImageUrl(item);
  console.log("Receipt Image URL:", imageUrl);
  console.log("Receipt Image Path:", item.image_path);

  dialog.value = true;
};

const isPDFReceipt = (receipt) => {
  return (
    receipt &&
    receipt.image_path &&
    receipt.image_path.toLowerCase().endsWith(".pdf")
  );
};

const getReceiptImageUrl = (receipt) => {
  if (!receipt) return "";
  return api.getReceiptImageUrl(receipt.id);
};

const saveReceipt = async () => {
  saving.value = true;
  try {
    // Update the receipt
    await api.updateReceipt(selectedReceipt.value.id, editedReceipt.value);

    // Reload the updated receipt from the server to get the new image_path
    const updatedReceipt = await api.getReceiptById(selectedReceipt.value.id);

    // Update the receipt in the local array with fresh data from server
    const index = receipts.value.findIndex(
      (r) => r.id === selectedReceipt.value.id
    );
    if (index !== -1) {
      receipts.value[index] = {
        ...updatedReceipt,
        date: new Date(updatedReceipt.date),
      };
    }

    // Recalculate totals
    amountAvailable.value = receipts.value
      .filter(
        (r) =>
          !r.used && (r.hsa_status === "Yes" || r.hsa_status === "Partially")
      )
      .reduce((sum, r) => sum + r.total_amount, 0);
    amountUsed.value = receipts.value
      .filter((r) => r.used)
      .reduce((sum, r) => sum + r.total_amount, 0);

    dialog.value = false;
  } catch (err) {
    alert(`Failed to save: ${err.message}`);
  } finally {
    saving.value = false;
  }
};

const deleteReceipt = () => {
  deleteDialog.value = true;
};

const confirmDelete = async () => {
  deleting.value = true;
  try {
    await api.deleteReceipt(selectedReceipt.value.id);

    receipts.value = receipts.value.filter(
      (r) => r.id !== selectedReceipt.value.id
    );

    amountAvailable.value = receipts.value
      .filter(
        (r) =>
          !r.used && (r.hsa_status === "Yes" || r.hsa_status === "Partially")
      )
      .reduce((sum, r) => sum + r.total_amount, 0);
    amountUsed.value = receipts.value
      .filter((r) => r.used)
      .reduce((sum, r) => sum + r.total_amount, 0);

    deleteDialog.value = false;
    dialog.value = false;
  } catch (err) {
    alert(`Failed to delete: ${err.message}`);
  } finally {
    deleting.value = false;
  }
};

onMounted(() => {
  loadReceipts();
});
</script>

<style scoped>
.receipt-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 300px;
  background-color: #f5f5f5;
  border-radius: 8px;
  padding: 16px;
}

.pdf-viewer {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.image-viewer {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.image-viewer .v-img {
  width: 100%;
  border: 1px solid #e0e0e0;
}
</style>
