# Design System
# SEVIMA-PAY: Multi-Tenant UKT Payment System

> **Version**: 1.0.0
> **Framework**: React 18/19 + Vite + Tailwind CSS + shadcn/ui
> **Language**: Bahasa Indonesia (primary), English (technical labels)

---

## 1. Design Principles

| Principle | Description |
|-----------|-------------|
| **Clarity** | Financial data harus mudah dibaca tanpa ambiguitas |
| **Trust** | Desain premium yang membangun kepercayaan pengguna |
| **Efficiency** | Admin keuangan bekerja dengan ratusan tagihan — UI harus cepat dan efisien |
| **Accessibility** | WCAG 2.1 AA minimum, mendukung screen reader |
| **Multi-tenant Awareness** | Identitas institusi jelas terlihat di setiap halaman |

---

## 2. Color System

### 2.1 Brand Colors

```css
/* Primary — Deep Blue (Trust, Finance) */
--primary-50: hsl(220, 70%, 96%);
--primary-100: hsl(220, 70%, 90%);
--primary-200: hsl(220, 70%, 80%);
--primary-300: hsl(220, 65%, 70%);
--primary-400: hsl(220, 65%, 58%);
--primary-500: hsl(220, 65%, 48%);  /* Main */
--primary-600: hsl(220, 70%, 40%);
--primary-700: hsl(220, 75%, 32%);
--primary-800: hsl(220, 78%, 24%);
--primary-900: hsl(220, 80%, 16%);
--primary-950: hsl(220, 85%, 10%);

/* Accent — Teal (Success, Action) */
--accent-50: hsl(172, 60%, 96%);
--accent-100: hsl(172, 60%, 88%);
--accent-200: hsl(172, 55%, 76%);
--accent-300: hsl(172, 50%, 62%);
--accent-400: hsl(172, 50%, 50%);
--accent-500: hsl(172, 55%, 40%);  /* Main */
--accent-600: hsl(172, 60%, 32%);
--accent-700: hsl(172, 65%, 24%);
```

### 2.2 Semantic Colors

```css
/* Status Colors — critical for payment system */
--status-unpaid: hsl(38, 92%, 50%);       /* Amber */
--status-unpaid-bg: hsl(38, 92%, 95%);
--status-unpaid-text: hsl(38, 92%, 30%);

--status-paid: hsl(142, 71%, 45%);         /* Green */
--status-paid-bg: hsl(142, 71%, 95%);
--status-paid-text: hsl(142, 71%, 25%);

--status-expired: hsl(0, 72%, 51%);        /* Red */
--status-expired-bg: hsl(0, 72%, 95%);
--status-expired-text: hsl(0, 72%, 30%);

/* Neutral */
--neutral-50: hsl(220, 14%, 96%);
--neutral-100: hsl(220, 13%, 91%);
--neutral-200: hsl(220, 12%, 84%);
--neutral-300: hsl(220, 10%, 72%);
--neutral-400: hsl(220, 8%, 56%);
--neutral-500: hsl(220, 8%, 42%);
--neutral-600: hsl(220, 10%, 34%);
--neutral-700: hsl(220, 12%, 26%);
--neutral-800: hsl(220, 16%, 18%);
--neutral-900: hsl(220, 20%, 12%);
--neutral-950: hsl(220, 25%, 6%);
```

### 2.3 Dark Mode

```css
/* Dark mode overrides — uses the same HSL values with adjusted lightness */
[data-theme="dark"] {
    --bg-primary: var(--neutral-950);
    --bg-secondary: var(--neutral-900);
    --bg-tertiary: var(--neutral-800);
    --text-primary: var(--neutral-50);
    --text-secondary: var(--neutral-300);
    --border-default: var(--neutral-700);
}
```

---

## 3. Typography

### 3.1 Font Stack

```css
/* Primary: Inter — excellent readability for data-heavy interfaces */
--font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;

/* Monospace: JetBrains Mono — for invoice numbers, amounts */
--font-mono: 'JetBrains Mono', 'Fira Code', 'Courier New', monospace;
```

### 3.2 Type Scale

| Token | Size | Weight | Line Height | Use Case |
|-------|------|--------|-------------|----------|
| `text-display` | 2.25rem (36px) | 700 | 1.2 | Page titles |
| `text-heading-1` | 1.875rem (30px) | 600 | 1.3 | Section headers |
| `text-heading-2` | 1.5rem (24px) | 600 | 1.35 | Card titles |
| `text-heading-3` | 1.25rem (20px) | 600 | 1.4 | Subsection headers |
| `text-body-lg` | 1.125rem (18px) | 400 | 1.6 | Emphasized body |
| `text-body` | 1rem (16px) | 400 | 1.6 | Default body |
| `text-body-sm` | 0.875rem (14px) | 400 | 1.5 | Table data, secondary info |
| `text-caption` | 0.75rem (12px) | 500 | 1.4 | Labels, metadata |
| `text-amount` | 1.25rem (20px) | 700 | 1.2 | Monetary amounts (mono) |

---

## 4. Spacing System

```css
/* 4px base unit */
--space-1: 0.25rem;   /* 4px */
--space-2: 0.5rem;    /* 8px */
--space-3: 0.75rem;   /* 12px */
--space-4: 1rem;      /* 16px */
--space-5: 1.25rem;   /* 20px */
--space-6: 1.5rem;    /* 24px */
--space-8: 2rem;      /* 32px */
--space-10: 2.5rem;   /* 40px */
--space-12: 3rem;     /* 48px */
--space-16: 4rem;     /* 64px */
```

---

## 5. Component Library

### 5.1 shadcn/ui (React) Components Used

| Component | Use Case | Customization |
|-----------|----------|---------------|
| `Button` | Actions, form submit | Primary, Secondary, Ghost, Destructive variants |
| `Card` | Dashboard stats, invoice detail | Custom header with status indicator |
| `Table` | Invoice list, transaction history | Sticky header, sortable columns |
| `Badge` | Invoice status | Custom colors per status |
| `Dialog` | Confirmation dialogs | Delete confirmation, invoice detail |
| `Input` | Form fields | With validation states |
| `Select` | Dropdowns, filters | Searchable for student selection |
| `Tabs` | Dashboard sections | Underline style |
| `Skeleton` | Loading states | Match component dimensions |
| `Toast` / `Sonner` | Notifications | Success, error, warning variants |
| `Sheet` | Mobile navigation | Slide-in sidebar |
| `DropdownMenu` | User menu, actions | With icons |
| `Pagination` | Table pagination | Page numbers + per-page selector |
| `DatePicker` | Due date selection | Calendar popup |
| `Separator` | Visual dividers | Horizontal rule |

### 5.2 Custom Components

#### StatusBadge

```tsx
// Consistent status display across the app
<StatusBadge status="paid" />
<StatusBadge status="unpaid" />
<StatusBadge status="expired" />
```

| Status | Color | Icon | Label |
|--------|-------|------|-------|
| `unpaid` | Amber | `Clock` | Belum Bayar |
| `paid` | Green | `CheckCircle` | Lunas |
| `expired` | Red | `XCircle` | Kedaluwarsa |

#### AmountDisplay

```tsx
// Formatted currency display with mono font
<AmountDisplay amount={5000000} />
// Output: Rp 5.000.000
```

#### InstitutionBadge

```tsx
// Shows institution context in multi-tenant views
<InstitutionBadge code="UNIV-NUS" name="Universitas Nusantara" />
```

#### EmptyState

```tsx
// Consistent empty state with illustration
<EmptyState
    title="Belum ada tagihan"
    description="Buat tagihan pertama untuk mahasiswa"
    action={{ label: "Buat Tagihan", href: "/admin/invoices/new" }}
/>
```

#### StatCard

```tsx
// Dashboard statistics card with trend
<StatCard
    title="Total Tagihan"
    value="Rp 150.000.000"
    subtitle="12 tagihan aktif"
    trend={{ direction: "up", value: "8%" }}
    icon={Receipt}
/>
```

---

## 6. Layout System

### 6.1 Page Layout Structure

```
┌─────────────────────────────────────────────────┐
│ Topbar (Institution name, User menu, Theme)     │
├────────┬────────────────────────────────────────┤
│        │                                        │
│ Sidebar│  Main Content Area                     │
│  Nav   │  ┌────────────────────────────────┐    │
│        │  │ Page Header (Title + Actions)  │    │
│ - Dash │  ├────────────────────────────────┤    │
│ - Bills│  │                                │    │
│ - ...  │  │  Content                       │    │
│        │  │                                │    │
│        │  │                                │    │
│        │  └────────────────────────────────┘    │
│        │                                        │
└────────┴────────────────────────────────────────┘
```

### 6.2 Breakpoints

| Breakpoint | Width | Layout |
|------------|-------|--------|
| `mobile` | < 640px | Single column, bottom nav |
| `tablet` | 640px - 1024px | Collapsed sidebar (hamburger) |
| `desktop` | > 1024px | Full sidebar + content |
| `wide` | > 1280px | Wider content area |

### 6.3 Grid System

```css
/* Content area max width */
--content-max-width: 1200px;

/* Sidebar width */
--sidebar-width: 260px;
--sidebar-collapsed-width: 64px;

/* Card grid */
.stat-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: var(--space-6);
}
```

---

## 7. Navigation Structure

### 7.1 Super Admin

```
📊 Dashboard
🏛️ Institusi
   ├── Daftar Institusi
   └── Tambah Institusi
👥 Pengguna
```

### 7.2 Admin Keuangan

```
📊 Dashboard
📋 Tagihan
   ├── Daftar Tagihan
   └── Buat Tagihan
🎓 Mahasiswa
📜 Riwayat Transaksi
```

### 7.3 Mahasiswa

```
📊 Dashboard
📋 Tagihan Saya
📜 Riwayat Pembayaran
```

---

## 8. Animations & Micro-interactions

### 8.1 Transitions

```css
/* Default transition */
--transition-fast: 150ms ease;
--transition-normal: 250ms ease;
--transition-slow: 350ms ease;

/* Page transitions */
--page-enter: 200ms ease-out;
--page-exit: 150ms ease-in;
```

### 8.2 Micro-interactions

| Element | Interaction | Animation |
|---------|-------------|-----------|
| Button | Hover | Scale 1.02, shadow increase |
| Button | Click | Scale 0.98, then bounce back |
| Card | Hover | Subtle shadow elevation |
| Badge status change | Update | Pulse animation (scale 1.0→1.1→1.0) |
| Table row | Hover | Background highlight |
| Sidebar item | Active | Left border accent + bg highlight |
| Toast | Enter | Slide in from right + fade |
| Toast | Exit | Slide out right + fade |
| Modal | Open | Scale 0.95→1.0 + backdrop fade |
| Skeleton | Loading | Shimmer gradient animation |
| Stat card value | Load | Count-up animation |

### 8.3 Loading States

```tsx
// Skeleton for stat cards
<div className="stat-grid">
    {Array.from({ length: 4 }).map((_, i) => (
        <Skeleton key={i} className="h-32 rounded-xl" />
    ))}
</div>

// Skeleton for table rows
<div className="space-y-2">
    {Array.from({ length: 5 }).map((_, i) => (
        <Skeleton key={i} className="h-12 rounded-md" />
    ))}
</div>
```

---

## 9. Icon System

**Library**: Lucide Icons (via `lucide-react`)

| Category | Icons |
|----------|-------|
| Navigation | `LayoutDashboard`, `Receipt`, `Users`, `Building2`, `History` |
| Actions | `Plus`, `Search`, `Filter`, `Download`, `RefreshCw` |
| Status | `CheckCircle`, `Clock`, `XCircle`, `AlertTriangle` |
| Auth | `LogIn`, `LogOut`, `User`, `Shield` |
| Utility | `ChevronDown`, `ChevronRight`, `MoreVertical`, `ExternalLink` |

---

## 10. Form Patterns

### 10.1 Validation Display

```tsx
// Inline validation with error message
<div className="form-group">
    <Label htmlFor="amount">Nominal Tagihan</Label>
    <Input
        id="amount"
        type="number"
        value={amount}
        onChange={(e) => setAmount(Number(e.target.value))}
        className={errors.amount ? 'border-destructive' : ''}
    />
    {errors.amount && (
        <p className="text-sm text-destructive mt-1">{errors.amount}</p>
    )}
</div>
```

### 10.2 Form Layout

- **Single column** for simple forms (login, single entity)
- **Two column** for complex forms (invoice creation with student + amount details)
- **Labels above inputs** (not inline) for readability
- **Required indicator**: red asterisk `*` after label
- **Help text**: small gray text below input

---

## 11. Data Display Patterns

### 11.1 Money Formatting

```typescript
// Always show full amount, no abbreviations
// Indonesian locale: period as thousands separator
formatCurrency(5000000) → "Rp 5.000.000"
formatCurrency(750000)  → "Rp 750.000"
```

### 11.2 Date Formatting

```typescript
// Indonesian format
formatDate("2026-12-31") → "31 Desember 2026"
formatDateTime("2026-12-31T10:30:00Z") → "31 Des 2026, 10:30 WIB"
formatRelative("2026-12-31T10:30:00Z") → "3 hari lagi" / "2 jam yang lalu"
```

### 11.3 Table Patterns

- **Sticky header** on scroll
- **Sortable columns** with sort indicator icon
- **Row hover highlight**
- **Status badge** in status column
- **Actions column** with dropdown menu (rightmost)
- **Responsive**: horizontal scroll on mobile with frozen first column
- **Empty state**: illustration + action button when no data
- **Pagination**: bottom of table with page count + per-page selector

---

## 12. Accessibility

| Requirement | Implementation |
|-------------|---------------|
| Color contrast | WCAG AA: 4.5:1 for text, 3:1 for large text |
| Keyboard navigation | All interactive elements focusable, visible focus ring |
| Screen reader | ARIA labels on icons, status announcements |
| Focus management | Trap focus in modals, return focus on close |
| Error identification | Error messages linked to inputs via `aria-describedby` |
| Status indicators | Never rely on color alone — icon + text |
