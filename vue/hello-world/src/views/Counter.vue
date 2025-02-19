<script setup>
import { ref } from 'vue'

const count = ref(0);

function increment() {
  count.value++;
}
</script>

<template>
  <div class="">
    <h1>Counter</h1>
    <div>count: {{ count }}</div>
    <button @click="increment">+</button>
  </div>
</template>

<style>
@media (min-width: 1024px) {
  .about {
    min-height: 100vh;
    display: flex;
    align-items: center;
  }
}
</style>
