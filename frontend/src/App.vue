<template>
  <v-app>
    <v-app-bar color="purple-darken-2" prominent>
      <v-app-bar-nav-icon @click="drawer = !drawer"></v-app-bar-nav-icon>
      <v-toolbar-title>HSA Receipt Manager</v-toolbar-title>
    </v-app-bar>

    <v-navigation-drawer v-model="drawer" temporary>
      <v-list>
        <v-list-item
          v-for="item in menuItems"
          :key="item.title"
          :to="item.to"
          link
        >
          <template v-slot:prepend>
            <v-icon>{{ item.icon }}</v-icon>
          </template>
          <v-list-item-title>{{ item.title }}</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>

    <v-main>
      <v-container fluid>
        <router-view></router-view>
      </v-container>
    </v-main>
  </v-app>
</template>

<script setup>
import { ref } from "vue";

const drawer = ref(false);

const menuItems = [
  { title: "Upload Receipt", icon: "mdi-cloud-upload-outline", to: "/upload" },
  { title: "My Receipts", icon: "mdi-receipt-text-outline", to: "/receipts" },
  { title: "HSA Deduction", icon: "mdi-calculator", to: "/deduct" },
];
</script>
