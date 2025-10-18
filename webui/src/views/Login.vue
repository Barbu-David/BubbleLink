<!-- src/views/Login.vue -->
<template>
  <section class="wrap">
    <div class="card">
      <h1>Join BubbleLink</h1>
      <p class="muted">
        Create a lightweight account to explore places and unlock location threads.
      </p>

      <form @submit.prevent="submit">
        <div class="row">
          <label>Username</label>
          <input v-model.trim="name" placeholder="3–16 chars" minlength="3" maxlength="16" required />
        </div>
        <div class="row grid">
          <div>
            <label>Country</label>
            <input v-model.trim="country" placeholder="Romania" required />
          </div>
          <div>
            <label>City</label>
            <input v-model.trim="city" placeholder="Bucuresti" required />
          </div>
        </div>

        <ErrorMsg v-if="error" :msg="error" />
        <button class="cta" :disabled="loading">
          <LoadingSpinner v-if="loading" />
          <span v-else>Continue</span>
        </button>
      </form>
    </div>
  </section>
</template>

<script setup>
import { ref, getCurrentInstance } from 'vue'
import { useRouter } from 'vue-router'

const { appContext } = getCurrentInstance()
const axios = appContext.config.globalProperties.$axios  // uses your axios wrapper with __API_URL__ 
const router = useRouter()

const name = ref('')
const country = ref('')
const city = ref('')

const loading = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const body = {
      Name: { name: name.value }, // backend reads Name.FormatedName as "name" field :contentReference[oaicite:4]{index=4}
      country: country.value,
      city: city.value
    }
    // The handler expects JSON fields exactly: { Name: { FormatedName: string }, Country, City }
    // We'll map correctly:
    const payload = {
      Name: { FormatedName: name.value },
      Country: country.value,
      City: city.value
    }
    const res = await axios.post('/session', payload)
    const { userId, apiKey } = res.data
    localStorage.setItem('bubble_user', JSON.stringify({ userId, apiKey, name: name.value, country: country.value, city: city.value }))
    //setUser({ userId, apiKey, name: name.value, country: country.value, city: city.value })
    router.push({ name: 'map' })
  } catch (e) {
    error.value = e?.response?.data?.error || 'Login failed. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.wrap {
  min-height: calc(100vh - 64px);
  display: grid; place-items: center; padding: 28px 16px;
}
.card {
  width: 100%; max-width: 520px;
  background: rgba(255,255,255,0.06); border: 1px solid rgba(255,255,255,0.1);
  border-radius: 16px; padding: 22px; box-shadow: 0 12px 40px rgba(0,0,0,.25);
}
h1 { margin: 0 0 6px; font-size: 28px; }
.muted { opacity: .7; margin-bottom: 18px; }
.row { display: grid; gap: 8px; margin-bottom: 12px; }
.grid { grid-template-columns: 1fr 1fr; gap: 12px; }
label { font-size: 12px; letter-spacing: .4px; opacity: .85; }
input {
  width: 100%; padding: 10px 12px; border-radius: 10px;
  border: 1px solid rgba(255,255,255,.15); background: rgba(0,0,0,.25); color: #fff;
}
.cta {
  width: 100%; margin-top: 10px;
  display: grid; place-items: center;
  background: linear-gradient(135deg, #5be8b2, #25c1ff);
  color: #0c0f1a; font-weight: 800; border: 0; border-radius: 12px; padding: 12px; cursor: pointer;
}
</style>
