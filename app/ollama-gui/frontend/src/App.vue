<template>
  <div class="container">
    <header>
      <div class="logo">
        <img src="./assets/ollama-logo.svg" alt="Ollama Logo" />
        <h1>Ollama</h1>
      </div>
    </header>

    <main>
      <div class="sidebar">
        <h2>Available Models</h2>
        <div v-if="loading" class="loading">Loading models...</div>
        <div v-else-if="models.length === 0" class="no-models">
          No models found. Use the button below to pull a model.
        </div>
        <ul v-else class="model-list">
          <li 
            v-for="model in models" 
            :key="model.name"
            :class="{ active: activeModel === model.name }"
            @click="selectModel(model.name)"
          >
            {{ model.name }}
          </li>
        </ul>
        <button class="pull-model-btn" @click="showPullModelDialog = true">Pull Model</button>
      </div>

      <div class="chat-container">
        <div v-if="!activeModel" class="select-model">
          <h2>Select a model to start chatting</h2>
        </div>
        <div v-else class="chat">
          <div class="chat-header">
            <h2>Chat with {{ activeModel }}</h2>
          </div>
          <div class="messages" ref="messagesContainer">
            <div v-for="(msg, index) in messages" :key="index" class="message" :class="msg.role">
              <div class="message-content">{{ msg.content }}</div>
            </div>
          </div>
          <div class="input-area">
            <textarea 
              v-model="userMessage" 
              placeholder="Type your message here..." 
              @keydown.enter.prevent="sendMessage"
            ></textarea>
            <button @click="sendMessage" :disabled="isProcessing">
              <span v-if="!isProcessing">Send</span>
              <span v-else>Processing...</span>
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- Pull Model Dialog -->
    <div v-if="showPullModelDialog" class="dialog-overlay">
      <div class="dialog">
        <h2>Pull a Model</h2>
        <div class="dialog-content">
          <p>Enter the name of the model you want to pull:</p>
          <input v-model="modelToPull" type="text" placeholder="e.g., llama3.2" />
          <div class="popular-models">
            <p>Popular models:</p>
            <div class="model-buttons">
              <button @click="modelToPull = 'llama3.2'">Llama 3.2</button>
              <button @click="modelToPull = 'phi4'">Phi 4</button>
              <button @click="modelToPull = 'mistral'">Mistral</button>
              <button @click="modelToPull = 'gemma2'">Gemma 2</button>
            </div>
          </div>
        </div>
        <div class="dialog-actions">
          <button @click="showPullModelDialog = false">Cancel</button>
          <button @click="pullModel" :disabled="!modelToPull || pullingModel">
            {{ pullingModel ? 'Pulling...' : 'Pull Model' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { GetModels, Chat } from '../wailsjs/go/main/App';

export default {
  data() {
    return {
      models: [],
      activeModel: null,
      messages: [],
      userMessage: '',
      loading: true,
      isProcessing: false,
      showPullModelDialog: false,
      modelToPull: '',
      pullingModel: false
    };
  },
  mounted() {
    this.fetchModels();
    // Poll for models every 5 seconds
    setInterval(this.fetchModels, 5000);
  },
  watch: {
    messages() {
      this.$nextTick(() => {
        if (this.$refs.messagesContainer) {
          this.$refs.messagesContainer.scrollTop = this.$refs.messagesContainer.scrollHeight;
        }
      });
    }
  },
  methods: {
    async fetchModels() {
      try {
        const models = await GetModels();
        this.models = models;
        this.loading = false;
      } catch (error) {
        console.error('Error fetching models:', error);
      }
    },
    selectModel(modelName) {
      this.activeModel = modelName;
      this.messages = [];
    },
    async sendMessage() {
      if (!this.userMessage.trim() || this.isProcessing) return;
      
      const userMsg = this.userMessage.trim();
      this.messages.push({ role: 'user', content: userMsg });
      this.userMessage = '';
      this.isProcessing = true;
      
      try {
        const response = await Chat(this.activeModel, userMsg);
        this.messages.push({ role: 'assistant', content: response });
      } catch (error) {
        console.error('Error in chat:', error);
        this.messages.push({ 
          role: 'assistant', 
          content: 'Sorry, I encountered an error processing your request.' 
        });
      } finally {
        this.isProcessing = false;
      }
    },
    async pullModel() {
      if (!this.modelToPull || this.pullingModel) return;
      
      this.pullingModel = true;
      // In a real implementation, you would call a Go function to pull the model
      // For now, let's simulate a delay and update the models list
      setTimeout(() => {
        this.fetchModels();
        this.pullingModel = false;
        this.showPullModelDialog = false;
        this.modelToPull = '';
      }, 2000);
    }
  }
};
</script>

<style scoped>
.container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  font-family: 'Inter', sans-serif;
}

header {
  padding: 1rem;
  background-color: #f5f5f5;
  border-bottom: 1px solid #e0e0e0;
}

.logo {
  display: flex;
  align-items: center;
}

.logo img {
  height: 40px;
  margin-right: 1rem;
}

.logo h1 {
  font-size: 1.5rem;
  font-weight: 600;
  color: #333;
}

main {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.sidebar {
  width: 250px;
  background-color: #f9f9f9;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #e0e0e0;
}

.sidebar h2 {
  margin-top: 0;
  margin-bottom: 1rem;
  font-size: 1.2rem;
}

.model-list {
  list-style: none;
  padding: 0;
  margin: 0;
  overflow-y: auto;
  flex: 1;
}

.model-list li {
  padding: 0.75rem 1rem;
  cursor: pointer;
  border-radius: 4px;
  margin-bottom: 0.5rem;
}

.model-list li:hover {
  background-color: #f0f0f0;
}

.model-list li.active {
  background-color: #e6f7ff;
  font-weight: 600;
}

.pull-model-btn {
  margin-top: 1rem;
  padding: 0.75rem;
  background-color: #1677ff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.pull-model-btn:hover {
  background-color: #0958d9;
}

.chat-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.select-model {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  color: #888;
}

.chat {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-header {
  padding: 1rem;
  border-bottom: 1px solid #e0e0e0;
}

.chat-header h2 {
  margin: 0;
  font-size: 1.2rem;
}

.messages {
  flex: 1;
  padding: 1rem;
  overflow-y: auto;
}

.message {
  margin-bottom: 1rem;
  max-width: 80%;
}

.message.user {
  margin-left: auto;
  background-color: #e6f7ff;
  padding: 0.75rem 1rem;
  border-radius: 1rem 1rem 0 1rem;
}

.message.assistant {
  margin-right: auto;
  background-color: #f5f5f5;
  padding: 0.75rem 1rem;
  border-radius: 1rem 1rem 1rem 0;
}

.input-area {
  padding: 1rem;
  display: flex;
  border-top: 1px solid #e0e0e0;
}

.input-area textarea {
  flex: 1;
  padding: 0.75rem;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  resize: none;
  height: 60px;
  margin-right: 0.5rem;
}

.input-area button {
  padding: 0 1rem;
  background-color: #1677ff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.input-area button:disabled {
  background-color: #d9d9d9;
  cursor: not-allowed;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
}

.dialog {
  background-color: white;
  padding: 1.5rem;
  border-radius: 8px;
  width: 400px;
}

.dialog h2 {
  margin-top: 0;
  margin-bottom: 1rem;
}

.dialog-content {
  margin-bottom: 1.5rem;
}

.dialog-content input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  margin-bottom: 1rem;
}

.popular-models p {
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.model-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.model-buttons button {
  padding: 0.5rem 1rem;
  background-color: #f0f0f0;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  cursor: pointer;
}

.model-buttons button:hover {
  background-color: #e6e6e6;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.dialog-actions button {
  padding: 0.75rem 1rem;
  border-radius: 4px;
  cursor: pointer;
}

.dialog-actions button:first-child {
  background-color: white;
  border: 1px solid #d9d9d9;
}

.dialog-actions button:last-child {
  background-color: #1677ff;
  color: white;
  border: none;
}

.loading, .no-models {
  padding: 1rem;
  color: #888;
  text-align: center;
}
</style>
