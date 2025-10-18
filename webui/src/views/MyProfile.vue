<template>
  <div class="page narrow">
    <h1>My Profile</h1>

    <div v-if="!user">
      <p>You are not logged in.</p>
    </div>

    <form v-else @submit.prevent="save" class="card">
      <label>Display name</label>
      <input
        v-model.trim="name"
        :maxlength="16"
        placeholder="Enter your name"
        class="input"
      />

      <p class="hint">3–16 characters.</p>

      <p v-if="error" class="error">{{ error }}</p>

      <div class="row">
        <button class="btn" :disabled="busy || !valid">
          {{ busy ? "Saving…" : "Save" }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import axios from "@/services/axios";
import { useAuth } from "@/composables/useAuth";

const { user, setUser } = useAuth();

const name = ref("");
const busy = ref(false);
const error = ref("");

const valid = computed(
  () => name.value.length >= 3 && name.value.length <= 16
);

// Load current name from backend on mount
onMounted(async () => {
  if (!user.value) return;

  try {
    const { data } = await axios.get(`/users/${user.value.userId}/name`);
    name.value = data?.name || "";

    // keep local header in sync if empty
    if (!user.value.name && name.value) {
      setUser({ ...user.value, name: name.value });
    }
  } catch (err) {
    console.error("Failed to fetch username", err);
  }
});

// Save new name
async function save() {
  error.value = "";
  if (!user.value) return;
  if (!valid.value) {
    error.value = "Name must be between 3 and 16 characters.";
    return;
  }

  busy.value = true;
  try {
    // backend expects { Name: { name: "..." } }
    await axios.put(`/users/${user.value.userId}/name`, {
      Name: { name: name.value }
    });

    // update local user so header shows the change
    setUser({ ...user.value, name: name.value });
  } catch (e) {
    error.value = e?.response?.data?.error || "Could not save name.";
  } finally {
    busy.value = false;
  }
}
</script>

<style scoped>
.page.narrow {
  max-width: 560px;
  margin: 24px auto;
  padding: 0 16px;
}
.card {
  display: grid;
  gap: 10px;
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.04);
}
label {
  font-weight: 700;
}
.input {
  padding: 10px 12px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.06);
  color: #e9ecf1;
}
.hint {
  opacity: 0.7;
  font-size: 0.9em;
}
.error {
  color: #ff9b9b;
}
.row {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}
.btn {
  background: linear-gradient(135deg, #7aaaff, #8bd0ff);
  color: #0c0f1a;
  border: 0;
  padding: 10px 14px;
  border-radius: 10px;
  font-weight: 700;
  cursor: pointer;
}
.btn:disabled {
  opacity: 0.6;
  cursor: default;
}
</style>
