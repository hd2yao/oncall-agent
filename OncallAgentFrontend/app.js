const API_BASE = "http://localhost:6871";
const STORAGE_KEY = "oncallgpt.history";

const elements = {
  app: document.getElementById("app"),
  sidebar: document.getElementById("sidebar"),
  toggleSidebar: document.getElementById("toggleSidebar"),
  newChatBtn: document.getElementById("newChatBtn"),
  clearChatBtn: document.getElementById("clearChatBtn"),
  historyList: document.getElementById("historyList"),
  messages: document.getElementById("messages"),
  typing: document.getElementById("typingIndicator"),
  prompt: document.getElementById("promptInput"),
  sendBtn: document.getElementById("sendBtn"),
  modeSelect: document.getElementById("modeSelect"),
  fileInput: document.getElementById("fileInput"),
  fileTriggerBtn: document.getElementById("fileTriggerBtn"),
  uploadBtn: document.getElementById("uploadBtn"),
  toastContainer: document.getElementById("toastContainer"),
};

const state = {
  conversations: [],
  currentId: null,
};

function loadConversations() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    state.conversations = raw ? JSON.parse(raw) : [];
  } catch (error) {
    console.error(error);
    state.conversations = [];
  }

  if (!state.conversations.length) {
    createConversation();
  } else {
    state.currentId = state.conversations[0].id;
  }
}

function saveConversations() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(state.conversations));
}

function createConversation() {
  const id = crypto.randomUUID();
  const conversation = {
    id,
    title: "新对话",
    createdAt: Date.now(),
    messages: [],
  };
  state.conversations.unshift(conversation);
  state.currentId = id;
  saveConversations();
}

function getCurrentConversation() {
  return state.conversations.find((item) => item.id === state.currentId);
}

function setConversationTitle(conversation) {
  const firstUserMsg = conversation.messages.find((msg) => msg.role === "user");
  if (firstUserMsg) {
    conversation.title = firstUserMsg.content.slice(0, 20);
  }
}

function renderHistory() {
  elements.historyList.innerHTML = "";
  state.conversations.forEach((conversation) => {
    const item = document.createElement("li");
    item.className = "history-item" +
      (conversation.id === state.currentId ? " active" : "");
    item.innerHTML = `
      <strong>${escapeHtml(conversation.title)}</strong>
      <small>${new Date(conversation.createdAt).toLocaleString()}</small>
    `;
    item.addEventListener("click", () => {
      state.currentId = conversation.id;
      renderMessages();
      renderHistory();
    });
    elements.historyList.appendChild(item);
  });
}

function renderMessages() {
  const conversation = getCurrentConversation();
  elements.messages.innerHTML = "";
  if (!conversation) return;
  conversation.messages.forEach((message) => {
    const messageEl = createMessageElement(message.role, message.content);
    elements.messages.appendChild(messageEl);
  });
  requestAnimationFrame(scrollToBottom);
}

function appendMessage(role, content) {
  const conversation = getCurrentConversation();
  if (!conversation) return null;
  const message = { role, content, createdAt: Date.now() };
  conversation.messages.push(message);
  if (role === "user" && conversation.messages.length <= 2) {
    setConversationTitle(conversation);
  }
  saveConversations();
  const messageEl = createMessageElement(role, content);
  elements.messages.appendChild(messageEl);
  requestAnimationFrame(scrollToBottom);
  renderHistory();
  return messageEl;
}

function updateMessageElement(messageEl, content) {
  const contentEl = messageEl.querySelector(".content");
  contentEl.innerHTML = renderMarkdown(content);
  enhanceCodeBlocks(contentEl);
}

function createMessageElement(role, content) {
  const messageEl = document.createElement("div");
  messageEl.className = `message ${role}`;
  const avatarLabel = role === "user" ? "你" : "AI";
  messageEl.innerHTML = `
    <div class="avatar">${avatarLabel}</div>
    <div class="bubble"><div class="content"></div></div>
  `;
  updateMessageElement(messageEl, content);
  return messageEl;
}

function showTyping(show) {
  elements.typing.classList.toggle("hidden", !show);
}

function scrollToBottom() {
  elements.messages.scrollTop = elements.messages.scrollHeight;
}

function renderMarkdown(text) {
  if (!window.marked) return escapeHtml(text);
  return marked.parse(text, { breaks: true });
}

function enhanceCodeBlocks(container) {
  if (!window.hljs) return;
  const codeBlocks = container.querySelectorAll("pre code");
  codeBlocks.forEach((code) => {
    if (code.dataset.enhanced) return;
    const rawText = code.textContent;
    code.dataset.raw = rawText;
    window.hljs.highlightElement(code);

    const pre = code.parentElement;
    pre.classList.add("code-block");
    const button = document.createElement("button");
    button.className = "copy-btn";
    button.textContent = "复制";
    button.addEventListener("click", async () => {
      try {
        await navigator.clipboard.writeText(code.dataset.raw || "");
        button.textContent = "已复制";
        button.classList.add("copied");
        setTimeout(() => {
          button.textContent = "复制";
          button.classList.remove("copied");
        }, 1200);
      } catch (error) {
        toast("复制失败", "error");
      }
    });
    pre.appendChild(button);

    const lines = code.innerHTML.split(/\n/);
    if (lines.length && lines[lines.length - 1].trim() === "") {
      lines.pop();
    }
    code.innerHTML = lines
      .map((line, index) => {
        const safeLine = line || "&#8203;";
        return `
          <span class="code-line">
            <span class="line-no">${index + 1}</span>
            <span class="line-text">${safeLine}</span>
          </span>
        `;
      })
      .join("");
    code.dataset.enhanced = "true";
  });
}

function toast(message, type = "info") {
  const toastEl = document.createElement("div");
  toastEl.className = `toast ${type}`;
  toastEl.textContent = message;
  elements.toastContainer.appendChild(toastEl);
  setTimeout(() => {
    toastEl.remove();
  }, 2800);
}

async function sendMessage() {
  const text = elements.prompt.value.trim();
  if (!text) return;
  elements.prompt.value = "";
  resizeTextarea();

  appendMessage("user", text);
  showTyping(true);

  const mode = elements.modeSelect.value;
  try {
    if (mode === "stream") {
      await sendStreamMessage(text);
    } else {
      await sendQuickMessage(text);
    }
  } catch (error) {
    console.error(error);
    toast("请求失败，请检查后端", "error");
  } finally {
    showTyping(false);
  }
}

async function sendQuickMessage(question) {
  const conversation = getCurrentConversation();
  if (!conversation) return;
  const res = await fetch(`${API_BASE}/api/chat`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      id: conversation.id,
      question,
    }),
  });

  if (!res.ok) {
    throw new Error("chat request failed");
  }

  const data = await res.json();
  const reply = extractReply(data);
  appendMessage("assistant", reply || "OK");
}

async function sendStreamMessage(question) {
  const conversation = getCurrentConversation();
  if (!conversation) return;

  const res = await fetch(`${API_BASE}/api/chat_stream`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      id: conversation.id,
      question,
    }),
  });

  if (!res.ok || !res.body) {
    throw new Error("stream request failed");
  }

  const assistantEl = appendMessage("assistant", "");
  let buffer = "";
  let content = "";
  const decoder = new TextDecoder();
  const reader = res.body.getReader();

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    const events = buffer.split("\n\n");
    buffer = events.pop() || "";

    events.forEach((event) => {
      event
        .split("\n")
        .filter((line) => line.startsWith("data:"))
        .forEach((line) => {
          const payload = line.replace(/^data:\s?/, "").trim();
          if (!payload || payload === "[DONE]") return;
          const chunk = parseStreamPayload(payload);
          if (chunk) {
            content += chunk;
            if (assistantEl) updateMessageElement(assistantEl, content);
            scrollToBottom();
          }
        });
    });
  }

  if (assistantEl) {
    const conversation = getCurrentConversation();
    if (conversation) {
      const lastMessage = conversation.messages[conversation.messages.length - 1];
      if (lastMessage && lastMessage.role === "assistant") {
        lastMessage.content = content;
        saveConversations();
      }
    }
  }
}

function parseStreamPayload(payload) {
  try {
    const data = JSON.parse(payload);
    return (
      data.delta ||
      data.content ||
      data.message ||
      data.text ||
      extractReply(data)
    );
  } catch (error) {
    return payload;
  }
}

function extractReply(data) {
  if (!data) return "";
  if (typeof data === "string") return data;
  if (data.data && typeof data.data === "object") {
    const nested =
      data.data.answer ||
      data.data.reply ||
      data.data.content ||
      data.data.text ||
      data.data.message;
    if (nested) return nested;
  }
  if (data.answer || data.reply || data.content || data.text) {
    return data.answer || data.reply || data.content || data.text;
  }
  if (data.message && data.message !== "OK") {
    return data.message;
  }
  return "";
}

function escapeHtml(text) {
  return String(text ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function resizeTextarea() {
  const el = elements.prompt;
  el.style.height = "auto";
  el.style.height = `${Math.min(el.scrollHeight, 200)}px`;
}

async function uploadFile(file) {
  if (!file) return;
  const formData = new FormData();
  formData.append("file", file);
  const res = await fetch(`${API_BASE}/api/upload`, {
    method: "POST",
    body: formData,
  });
  if (!res.ok) {
    throw new Error("upload failed");
  }
  const data = await res.json();
  toast("上传成功", "success");
  appendMessage(
    "assistant",
    data.message || data.result || "文件已上传，可继续对话。"
  );
}

async function runAiOps(op) {
  const res = await fetch(`${API_BASE}/api/ai_ops`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ op }),
  });

  if (!res.ok) {
    throw new Error("ai ops failed");
  }

  const data = await res.json();
  appendMessage(
    "assistant",
    data.result || data.message || data.content || JSON.stringify(data)
  );
}

function bindEvents() {
  elements.toggleSidebar.addEventListener("click", () => {
    const collapsed = elements.sidebar.dataset.collapsed === "true";
    const next = collapsed ? "false" : "true";
    elements.sidebar.dataset.collapsed = next;
    elements.app.classList.toggle("collapsed", next === "true");
  });

  elements.newChatBtn.addEventListener("click", () => {
    createConversation();
    renderHistory();
    renderMessages();
  });

  elements.clearChatBtn.addEventListener("click", () => {
    const conversation = getCurrentConversation();
    if (!conversation) return;
    conversation.messages = [];
    saveConversations();
    renderMessages();
  });

  elements.sendBtn.addEventListener("click", sendMessage);

  elements.prompt.addEventListener("input", resizeTextarea);
  elements.prompt.addEventListener("keydown", (event) => {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      sendMessage();
    }
  });

  elements.fileTriggerBtn.addEventListener("click", () => {
    elements.fileInput.click();
  });

  elements.uploadBtn.addEventListener("click", () => {
    elements.fileInput.click();
  });

  elements.fileInput.addEventListener("change", async (event) => {
    const file = event.target.files[0];
    if (!file) return;
    try {
      await uploadFile(file);
    } catch (error) {
      toast("上传失败", "error");
    } finally {
      event.target.value = "";
    }
  });

  document.querySelectorAll("[data-op]").forEach((button) => {
    button.addEventListener("click", async () => {
      const op = button.dataset.op;
      try {
        toast(`AI Ops：${button.textContent}`, "info");
        await runAiOps(op);
      } catch (error) {
        toast("AI Ops 执行失败", "error");
      }
    });
  });
}

function init() {
  loadConversations();
  renderHistory();
  renderMessages();
  bindEvents();
  resizeTextarea();
}

init();
