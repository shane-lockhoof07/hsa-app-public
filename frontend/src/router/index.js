import { createRouter, createWebHistory } from "vue-router";
import UploadReceipt from "../views/UploadReceipt.vue";
import ReceiptList from "../views/ReceiptList.vue";
import HSADeduction from "../views/HSADeduction.vue";

const routes = [
  {
    path: "/",
    redirect: "/upload",
  },
  {
    path: "/upload",
    name: "Upload",
    component: UploadReceipt,
  },
  {
    path: "/receipts",
    name: "Receipts",
    component: ReceiptList,
  },
  {
    path: "/deduct",
    name: "Deduct",
    component: HSADeduction,
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL || "/"),
  routes,
});

export default router;
