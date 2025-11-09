import { defineStore } from "pinia";
import api from "../services/api";

export const useReceiptsStore = defineStore("receipts", {
  state: () => ({
    receipts: [],
    loading: false,
    error: null,
    initialized: false,
  }),

  getters: {
    amountAvailable: (state) => {
      return state.receipts
        .filter((r) => !r.used && r.hsa_qualified)
        .reduce((sum, r) => sum + r.total_amount, 0);
    },

    amountUsed: (state) => {
      return state.receipts
        .filter((r) => r.used)
        .reduce((sum, r) => sum + r.total_amount, 0);
    },

    eligibleReceipts: (state) => {
      return state.receipts.filter((r) => !r.used && r.hsa_qualified);
    },

    usedReceipts: (state) => {
      return state.receipts.filter((r) => r.used);
    },
  },

  actions: {
    async fetchReceipts() {
      this.loading = true;
      this.error = null;

      try {
        this.receipts = await api.getReceipts();
        this.initialized = true;
      } catch (error) {
        this.error = `Failed to load receipts: ${error.message}`;
        console.error("Failed to fetch receipts:", error);
      } finally {
        this.loading = false;
      }
    },

    async updateReceipt(id, updates) {
      try {
        const updatedReceipt = await api.updateReceipt(id, updates);

        // Update in local state
        const index = this.receipts.findIndex((r) => r.id === id);
        if (index !== -1) {
          this.receipts[index] = {
            ...this.receipts[index],
            ...updates,
            date: new Date(updates.date),
          };
        }

        return updatedReceipt;
      } catch (error) {
        this.error = `Failed to update receipt: ${error.message}`;
        throw error;
      }
    },

    async deleteReceipt(id) {
      try {
        await api.deleteReceipt(id);

        // Remove from local state
        this.receipts = this.receipts.filter((r) => r.id !== id);
      } catch (error) {
        this.error = `Failed to delete receipt: ${error.message}`;
        throw error;
      }
    },

    async markReceiptsAsUsed(receiptIds, useReason = null) {
      try {
        // Update each receipt
        for (const id of receiptIds) {
          await this.updateReceipt(id, {
            used: true,
            use_reason: useReason,
          });
        }
      } catch (error) {
        this.error = `Failed to mark receipts as used: ${error.message}`;
        throw error;
      }
    },

    // Call this after upload completes to refresh data
    async refreshAfterUpload() {
      await this.fetchReceipts();
    },

    clearError() {
      this.error = null;
    },
  },
});
