# Cartography Protocol & Process (製圖流程規範)

> **目的**：規範 AgentPlaybook 協作體系中，由 **Planner**（亦可應 Navigator 請求發起）委託專職製圖師 **Cartographer** 產出出版級架構圖表的標準非同步協作流程，徹底解決 HTML/SVG 代碼膨脹對主流程 Context Window 的污染問題。

> [Notice] Cartographer requires the diagram-design skill (https://github.com/cathrynlavery/diagram-design) for publication-grade diagram rendering.

---

## 1. 核心哲學與架構原則 (Core Philosophy)

### 1.1 What 與 How 的嚴格分離 (Separation of Concerns)
- **委託端 (Planner，亦可應 Navigator 請求發起) 決定「畫什麼 (What)」**：
  - 專注於業務架構、模組職責、因果順序與核心訊息。
  - **只輸出高層級的語意結構**（實體、關係動詞、分區邊界），絕不撰寫任何 SVG 幾何、CSS 樣式或 HTML 標籤。
- **製圖師 (Cartographer) 決定「怎麼畫 (How)」**：
  - 專注於視覺類型適配（39 種 Visual Types）、幾何排版、直角折線公式（Orthogonal elbow `r=8`）、4px 網格與焦點色彩分配。
  - 全權負責把語意轉化為自包含、出版級的 HTML/SVG 檔案。

### 1.2 Context 零污染原則 (Zero Context Pollution)
- HTML + Inline SVG 的標記代碼極其肥大（動輒 5,000 ~ 15,000 tokens）。
- Cartographer 在**獨立 Context Window（專屬 Pane 或獨立 Subagent）**中運作：
  - 所有的 HTML/SVG 生成、幾何微調、座標計算與規範檢查**完全封裝在 Cartographer 內部**。
  - 回傳給主流程的**僅有 1 行檔案路徑與極簡摘要（< 100 tokens）**，主流程 Context 保持 100% 純淨。

### 1.3 非同步解耦與無阻塞 (Fire-and-Forget Asynchrony)
- **委託端發射後不管**：Planner 派發 Brief 後，立即回歸核心主線任務（架構規劃、程式碼審查、使用者導讀），無需輪詢監控。
- **無狀態工匠**：Cartographer 畫完即落盤，完成後發送非同步輕量通知。

---

## 2. 標準五步協作流 (The 5-Step Cartography Flow)

```
[Step 1] Planner 識別視覺化需求 (定錨溝通目的；Navigator 可向 Planner 請求發起)
   │
   ▼
[Step 2] Planner 發送結構化繪圖摘要 (Structured Diagram Brief, ~150-200 tokens)
   │
   ▼
[Step 3] Cartographer 審美把關 (Taste Gate 評估：類型挑選 / 預算檢查 / 必要時建議)
   │
   ▼
[Step 4] Cartographer 獨立作畫 (4px 格線幾何渲染、執行 self_check.py 與驗證)
   │
   ▼
[Step 5] 落盤儲存並回報完成 (持久化至 docs/diagrams/*.html，發送非同步通知)
```

### 詳細階段說明：

| 步驟 | 負責角色 | 核心動作 | 輸出物 / 交付物 |
|---|---|---|---|
| **Step 1: 需求識別** | Planner *(註：Navigator 可請求 Planner 發起，但 Planner 為唯一標準 Actor，嚴格遵循 flows.json 不允許 Navigator 擔任 flow actor 規範)* | 評估讀者是否能從圖表中獲得高於純文字的理解價值。明確要傳達的架構重點與雙重門禁。 | 視覺化意圖定義 |
| **Step 2: 結構化派發** | Planner | 撰寫極簡 Diagram Brief，包含主題、目標檔案路徑、實體角色、流轉階段與底部要點。 | `[Cartographer Request]` (~200 tokens) |
| **Step 3: 審美把關** | Cartographer | 執行 `/diagram-design` 規範評估：<br>1. 複雜度預算檢視（≤ 12 節點）。<br>2. 形式檢視（若 3 欄表格或段落更優，主動建言免畫）。<br>3. 從 39 種視覺類型中挑選最佳語意呈現（如 Sequence, Architecture, Process）。 | 視覺決策 / 專業建議反饋 |
| **Step 4: 幾何渲染** | Cartographer | 1. 載入對應型態規範（`type-*.md`）與範本。<br>2. 計算座標、直角圓弧（`r=8`）、遮罩留白（6-10px）。<br>3. 執行幾何與無障礙腳本驗證（`self_check.py`, `verify-geometry.py`）。 | 自包含 HTML 檔案 |
| **Step 5: 落盤回報** | Cartographer | 將圖表持久化至目標路徑（如 `docs/diagrams/<slug>.html`）。向委託端回報單行完成資訊。 | `docs/diagrams/*.html` (零污染交付) |

---

## 3. 輸入契約規格書 (Diagram Brief Specification)

委託端發送給 Cartographer 的訊息必須符合標準化的結構格式，嚴格控制在 200 tokens 左右：

```markdown
[Cartographer Request: <主題簡稱>]
主題: <完整圖表標題>
目標儲存路徑: docs/diagrams/<slug>.html

溝通目的:
<簡述 1~3 點本圖表要傳達的核心架構概念、邊界約束或關鍵門禁>

語意資料與互動流程 (What):
• 角色 / 實體 (Entities):
  - <角色名稱> (<實體類型: Focal/Backend/Store/User...>): <簡短職責與 Sublabel>
• 核心分區 (Zones / Containers, 選填):
  - <分區名稱>: 包含哪些角色
• 核心流轉階段 (Stages / Transitions):
  - <起點> -> <終點>: <大寫動作標籤 ≤14字> (<連線樣式: Default/Accent/Link-blue/Dashed>)
  - ... (或以 10-15 行標準 Mermaid 語法提供)

視覺決策 (How to draw):
由 Cartographer 全權評估挑選最適視覺類型（如 Sequence, Architecture, Process, Data Flow），並在 4/10 密度與複雜度預算內進行幾何排版。

底部卡片重點 (Summary Takeaways):
1. <重點一：架構或安全邊界>
2. <重點二：關鍵門禁或流轉保證>
3. <重點三：協作或拓撲特徵>
```

---

## 4. 製圖師品質與審美標準 (Taste Gate Standards)

Cartographer 產出圖表時必須嚴格符合 `/diagram-design` 核心規範：

1. **4/10 密度與複雜度預算**：
   - 核心節點數 ≤ 9（`balanced` 模式下最多 12 個）。
   - 箭頭連線 ≤ 12 條。
   - 超過預算必須主動拆解為「總覽圖 + 細節圖（Overview + Detail）」。
2. **六大連接線幾何硬約束**：
   - **強制直角圓弧**：所有非共軸節點連線必須採直角折線（`r = 8`），嚴禁斜向對角線。
   - **6–10px 懸空淨距**：連線標籤必須有不透明遮罩，且遮罩與線條本體必須維持 6–10px 淨距，嚴禁文字壓線。
   - **零重疊連線**：多條平行線間距 ≥ 12px；交叉處使用 Bridge/Hop 跨線基語。
   - **同邊扇出（Fan-out）**：同側多條出入口接點均勻分散（間距 ≥ 12px），不可共用同一出入點。
   - **禁止穿透非端點方塊**：線條不可穿越無關方塊（除不可避免之橫向跨區線，且須改為虛線 Transit）。
   - **遮罩防覆蓋**：連線標籤遮罩不可被後續渲染的節點裁切。
3. **色彩與字體階層**：
   - **單圖焦點原則**：Coral 珊瑚紅焦點（`accent`）全圖嚴格限制 ≤ 1~2 處。
   - **字體規範**：標題使用 `Instrument Serif`；節點主標使用 `Geist` (sans 12px/600)；技術標籤與動詞使用 `Geist Mono` (8-9px)；嚴禁使用 JetBrains Mono。
4. **Accessible SVG 契約**：
   - `<svg>` 必須標註 `role="img"` 與 `aria-labelledby`。
   - 第一個子元素為 `<title>` 與 `<desc>`，且 ID 必須加上圖表專屬前綴（避免跨圖衝突）。
5. **自我驗證流程**：
   - 產出檔案必須通過 `self_check.py` 與 `verify-geometry.py` 檢查。

---

## 5. 實戰案例覆盤：程式碼開發生命週期圖

在 `v0.3.1` 開發生命週期繪圖實戰中，本協議達成的效益如下：

- **輸入 Token 消耗**：僅約 **185 tokens**（純語意 Brief）。
- **隔離代碼量**：**30KB 自包含 HTML/SVG** 完全由 Cartographer 內部消化，主流程 Context Window 污染率為 **0%**。
- **Cartographer 自主視覺創新**：
  - 自行評估後採納 **`Sequence`（時序圖）** 配合 **`Full Editorial` 範本**。
  - 創新設計「幾何中介排列（`User -> Reviewer -> Planner -> Builder`）」，透過將 Planner 置中，直觀呈現 Reviewer 與 Builder 的零物理連線。
  - 自主建構「X = 820 氣閘隔離牆（Airlock Boundary）」實體化盲審防護牆（The Blind Barrier）。
  - 將 Gate 2 (`REVIEW_PASS`) 作為全圖唯一的 Coral 焦點高亮。
- **產出檔案**：[`docs/diagrams/code-development-lifecycle.html`](./code-development-lifecycle.html)
