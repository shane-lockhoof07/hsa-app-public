import axios from "axios";

// Dynamic API URL based on where the frontend is accessed from
const getApiUrl = () => {
  const configuredUrl = import.meta.env.VITE_API_URL;

  if (configuredUrl) {
    console.log("Using configured API URL:", configuredUrl);
    return configuredUrl;
  }

  // Dynamically construct API URL based on current host
  const protocol = window.location.protocol;
  const hostname = window.location.hostname;
  const apiUrl = `${protocol}//${hostname}:30081/api`;

  console.log("Constructed API URL:", apiUrl);
  console.log("Current location:", window.location.href);

  return apiUrl;
};

const API_URL = getApiUrl();

console.log("Final API_URL:", API_URL);

export default {
  async uploadReceipt(file) {
    console.log("Uploading to:", `${API_URL}/receipts/upload`);
    console.log("File details:", {
      name: file.name,
      type: file.type,
      size: file.size,
    });

    const formData = new FormData();
    formData.append("file", file);

    try {
      const response = await axios.post(
        `${API_URL}/receipts/upload`,
        formData,
        {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        }
      );
      return response.data;
    } catch (error) {
      console.error("Upload error details:", {
        message: error.message,
        response: error.response?.data,
        status: error.response?.status,
      });

      // Handle duplicate errors with specific messages
      if (error.response && error.response.status === 409) {
        const data = error.response.data;
        if (data.error === "duplicate") {
          throw new Error(data.message || "Duplicate receipt detected");
        }
      }
      throw error;
    }
  },

  async getReceipts() {
    console.log("Fetching receipts from:", `${API_URL}/receipts`);
    try {
      const response = await axios.get(`${API_URL}/receipts`);
      return response.data || [];
    } catch (error) {
      console.error("Get receipts error:", error);
      throw error;
    }
  },

  async updateReceipt(id, data) {
    console.log("Updating receipt:", id, data);
    try {
      const response = await axios.put(`${API_URL}/receipts/${id}`, data);
      return response.data;
    } catch (error) {
      console.error("Update receipt error:", error);
      throw error;
    }
  },

  async deleteReceipt(id) {
    console.log("Deleting receipt:", id);
    try {
      const response = await axios.delete(`${API_URL}/receipts/${id}`);
      return response.data;
    } catch (error) {
      console.error("Delete receipt error:", error);
      throw error;
    }
  },

  async calculateDeduction(amount) {
    console.log("Calculating deduction for amount:", amount);
    try {
      const response = await axios.post(`${API_URL}/receipts/deduct`, {
        amount: amount,
      });
      return response.data;
    } catch (error) {
      console.error("Calculate deduction error:", error);
      throw error;
    }
  },

async approveDeduction(receiptIds, useReason) {
  console.log("Approving deduction for receipts:", receiptIds, "Reason:", useReason);
  try {
    const promises = receiptIds.map((id) => {
      const payload = {
        used: true,
        use_reason: useReason,
      };
      console.log(`Sending PUT request for receipt ${id} with payload:`, payload);
      return axios.put(`${API_URL}/receipts/${id}`, payload);
    });
    
    const results = await Promise.all(promises);
    console.log("All receipts updated:", results.map(r => r.data));
    
    return results;
  } catch (error) {
    console.error("Approve deduction error:", error);
    throw error;
  }
},

  getReceiptImageUrl(receiptId) {
    // Receipt files are served from /receipts/file/{id}
    // Use the same base as API but without /api suffix
    const baseUrl = API_URL.replace("/api", "");
    const imageUrl = `${baseUrl}/receipts/file/${receiptId}`;
    console.log("Receipt image URL:", imageUrl);
    return imageUrl;
  },


  async getReceiptById(id) {
    console.log("Fetching receipt by ID:", id);
    try {
      const response = await axios.get(`${API_URL}/receipts/${id}`);
      return response.data;
    } catch (error) {
      console.error("Get receipt by ID error:", error);
      throw error;
    }
  },
};
