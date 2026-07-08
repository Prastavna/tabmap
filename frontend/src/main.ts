import {createApp} from 'vue'
import App from './App.vue'
import './style.css';
import ui from '@nuxt/ui/vue-plugin'

const app = createApp(App)

// Configure Nuxt UI with light theme
app.use(ui, {
  theme: {
    colorMode: {
      preference: 'light',
      fallback: 'light'
    }
  }
})

app.mount('#app')
