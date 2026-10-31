# Frontend Web Specification
# SEVIMA-PAY: React + Vite Frontend

> **Version**: 1.0.0  
> **Framework**: React 18/19 + Vite (TypeScript)  
> **Styling**: Tailwind CSS + shadcn/ui (Radix UI primitives)  
> **State Management**: TanStack Query (React Query) + Zustand  
> **Routing**: React Router v6/v7  
> **Architecture**: Clean Architecture  

---

## 1. Technology Stack

| Technology | Version | Purpose |
|-----------|---------|---------|
| React | 18 / 19 | UI component library & ecosystem standard SEVIMA |
| Vite | 5.x / 6.x | Fast build tool & dev server |
| TypeScript | 5.x | End-to-end type safety |
| React Router | v6 / v7 | Client-side routing with route guards & lazy loading |
| Tailwind CSS | v3 / v4 | Utility-first CSS & responsive design system |
| shadcn/ui | Latest | Accessible Radix UI based components |
| Lucide React | Latest | Modern icon set |
| TanStack Query | v5 | Server state management, caching, background refetching |
| Zustand | 4.x / 5.x | Client state (auth session, active tenant, UI preferences) |
| React Hook Form + Zod | Latest | Form handling & schema validation |
| Axios | Latest | HTTP client with automatic JWT refresh interceptor |
| date-fns | Latest | Date formatting with Indonesian locale (`id`) |

---

## 2. Directory Structure (Clean Architecture)

```
sevima-fe/
├── public/
│   └── favicon.ico
├── src/
│   ├── assets/                     # Images, illustrations, logos
│   ├── core/                       # Core abstractions & config
│   │   ├── config/                 # Environment variables & constants
│   │   ├── router/                 # React Router definitions & protected routes
│   │   │   ├── AppRouter.tsx
│   │   │   └── ProtectedRoute.tsx  # Role-based & tenant route guards
│   │   └── http/                   # Axios instance & interceptors
│   │       ├── apiClient.ts
│   │       └── tokenManager.ts
│   ├── domain/                     # Layer 1: Entities & Value Objects (Models)
│   │   ├── models/
│   │   │   ├── auth.model.ts
│   │   │   ├── institution.model.ts
│   │   │   ├── student.model.ts
│   │   │   ├── invoice.model.ts
│   │   │   └── webhook.model.ts
│   │   └── repositories/           # Repository interfaces
│   │       ├── IAuthRepository.ts
│   │       ├── IInstitutionRepository.ts
│   │       └── IInvoiceRepository.ts
│   ├── infrastructure/             # Layer 3: External Implementations (API Adapters)
│   │   └── repositories/
│   │       ├── AuthRepository.ts
│   │       ├── InstitutionRepository.ts
│   │       └── InvoiceRepository.ts
│   ├── application/                # Layer 2: Use Cases & Query Hooks
│   │   ├── auth/
│   │   │   ├── useLogin.ts
│   │   │   └── useAuthStore.ts     # Zustand store for JWT & user role
│   │   ├── institution/
│   │   │   ├── useInstitutions.ts
│   │   │   └── useCreateInstitution.ts
│   │   ├── invoice/
│   │   │   ├── useInvoices.ts
│   │   │   ├── useInvoiceDetail.ts
│   │   │   └── useCreateInvoice.ts
│   │   ├── student/
│   │   │   └── useKrsEligibility.ts
│   │   └── mockgateway/
│   │       └── useSimulatePayment.ts
│   └── presentation/               # Layer 4: UI Components & Views
│       ├── components/             # Reusable UI primitives (shadcn/ui)
│       │   ├── ui/
│       │   │   ├── button.tsx
│       │   │   ├── card.tsx
│       │   │   ├── dialog.tsx
│       │   │   ├── input.tsx
│       │   │   ├── select.tsx
│       │   │   ├── badge.tsx
│       │   │   ├── table.tsx
│       │   │   └── toast.tsx
│       │   └── shared/
│       │       ├── Navbar.tsx
│       │       ├── Sidebar.tsx
│       │       ├── StatusBadge.tsx
│       │       ├── CurrencyDisplay.tsx
│       │       └── KrsEligibilityBanner.tsx
│       ├── layouts/
│       │   ├── AuthLayout.tsx
│       │   ├── DashboardLayout.tsx
│       │   └── StudentLayout.tsx
│       └── pages/                  # Page-level components
│           ├── auth/
│           │   └── LoginPage.tsx
│           ├── superadmin/
│           │   ├── InstitutionListPage.tsx
│           │   └── AuditLogPage.tsx
│           ├── admin/
│           │   ├── InvoiceListPage.tsx
│           │   ├── CreateInvoicePage.tsx
│           │   ├── InvoiceDetailPage.tsx
│           │   └── ReconciliationPage.tsx
│           ├── student/
│           │   ├── StudentInvoiceListPage.tsx
│           │   ├── StudentInvoiceDetailPage.tsx
│           │   └── KrsStatusPage.tsx
│           └── mockgateway/
│               └── SimulatorPage.tsx # Mock payment sandbox with HMAC debug
├── index.html
├── package.json
├── tailwind.config.ts
├── tsconfig.json
└── vite.config.ts
```

---

## 3. Pages & Features Specification

### 3.1 Login Page (`/login`)
- **Access**: Public
- **Features**:
  - Email & Password input with React Hook Form + Zod schema validation.
  - Role selection or automatic redirect based on decoded JWT claims:
    - `super_admin` → `/superadmin/institutions`
    - `admin_keuangan` → `/admin/invoices`
    - `mahasiswa` → `/student/invoices`
  - Demo quick-fill buttons for fast testing/evaluation (Super Admin, Admin UGM, Mahasiswa UGM).

```tsx
// src/presentation/pages/auth/LoginPage.tsx
import React from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useLogin } from '@/application/auth/useLogin';
import { Card, CardHeader, CardTitle, CardContent } from '@/presentation/components/ui/card';
import { Button } from '@/presentation/components/ui/button';
import { Input } from '@/presentation/components/ui/input';

const loginSchema = z.object({
  email: z.string().email('Format email tidak valid'),
  password: z.string().min(6, 'Password minimal 6 karakter'),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export const LoginPage: React.FC = () => {
  const { mutate: login, isPending, error } = useLogin();
  const { register, handleSubmit, formState: { errors }, setValue } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = (data: LoginFormValues) => {
    login(data);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-900 p-4">
      <Card className="w-full max-w-md shadow-xl border-slate-200">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-2">
            <span className="font-bold text-2xl text-blue-600">SEVIMA<span className="text-emerald-500">-PAY</span></span>
          </div>
          <CardTitle className="text-xl font-bold">Portal Tagihan UKT Multi-Kampus</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div>
              <label className="text-sm font-medium">Email Institusi / Mahasiswa</label>
              <Input {...register('email')} placeholder="nama@kampus.ac.id" className="mt-1" />
              {errors.email && <p className="text-xs text-red-500 mt-1">{errors.email.message}</p>}
            </div>
            <div>
              <label className="text-sm font-medium">Kata Sandi</label>
              <Input type="password" {...register('password')} placeholder="••••••••" className="mt-1" />
              {errors.password && <p className="text-xs text-red-500 mt-1">{errors.password.message}</p>}
            </div>
            {error && <div className="p-3 bg-red-50 border border-red-200 text-red-700 rounded-md text-sm">{error.message}</div>}
            <Button type="submit" className="w-full bg-blue-600 hover:bg-blue-700" disabled={isPending}>
              {isPending ? 'Memproses...' : 'Masuk ke Sistem'}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
};
```

---

### 3.2 Mahasiswa Dashboard & KRS Eligibility (`/student/invoices`)
- **Access**: `mahasiswa` only.
- **Header**: Nama Mahasiswa, NIM, Program Studi, Nama Kampus, dan **Banner Status Kelayakan KRS**.
- **KRS Status Logic**:
  - Jika seluruh tagihan semester aktif berstatus `paid` → **Badge Hijau: "Eligible / Siap Pengisian KRS"**.
  - Jika ada tagihan `unpaid` atau `expired` → **Badge Merah: "Terkunci / Belum Lunas UKT"** dengan rincian tunggakan.
- **Fitur Interaktif**: Tombol "Bayar Sekarang" yang langsung membuka Simulator Gateway (Mock) untuk live-demo kelunasan dan verifikasi seketika perubahan status KRS menjadi Eligible.

```tsx
// src/presentation/components/shared/KrsEligibilityBanner.tsx
import React from 'react';
import { CheckCircle2, AlertTriangle } from 'lucide-react';
import { Card, CardContent } from '@/presentation/components/ui/card';

interface KrsEligibilityProps {
  isEligible: boolean;
  unpaidCount: number;
  activeSemester: string;
}

export const KrsEligibilityBanner: React.FC<KrsEligibilityProps> = ({ isEligible, unpaidCount, activeSemester }) => {
  return (
    <Card className={`border-l-4 ${isEligible ? 'border-l-emerald-500 bg-emerald-50 dark:bg-emerald-950/20' : 'border-l-amber-500 bg-amber-50 dark:bg-amber-950/20'}`}>
      <CardContent className="p-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          {isEligible ? (
            <CheckCircle2 className="w-7 h-7 text-emerald-600 flex-shrink-0" />
          ) : (
            <AlertTriangle className="w-7 h-7 text-amber-600 flex-shrink-0" />
          )}
          <div>
            <h4 className="font-semibold text-slate-900 dark:text-slate-100">
              {isEligible ? 'Status KRS: Memenuhi Syarat (Eligible)' : 'Status KRS: Ditangguhkan (Belum Lunas)'}
            </h4>
            <p className="text-sm text-slate-600 dark:text-slate-300">
              {isEligible
                ? `Tagihan UKT semester ${activeSemester} telah lunas. Anda dapat melanjutkan pengisian KRS di SIAKAD.`
                : `Terdapat ${unpaidCount} tagihan UKT aktif yang belum terselesaikan. Pelunasan wajib sebelum KRS dibuka.`}
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
```

---

### 3.3 Admin Keuangan Dashboard (`/admin/invoices`)
- **Access**: `admin_keuangan` only (otomatis terisolasi pada `institusi_id` akun admin).
- **Fitur Utama**:
  1. **Statistik Real-time**: Total Tagihan Terbit, Total Lunas, Total Tunggakan, Total Expired.
  2. **Tabel Tagihan**: Filter status (`unpaid`, `paid`, `expired`), semester, tahun akademik, search NIM/nama.
  3. **Buat Tagihan Baru** (`/admin/invoices/new`): Form input tagihan individual atau upload CSV massal.
  4. **Rekonsiliasi & Audit Log**: Riwayat status perubahan webhook secara mendalam.

---

### 3.4 Simulator Payment Gateway (`/mock-gateway`)
- **Fitur Live Coding & Demo**:
  - Halaman simulator khusus untuk menirukan respons Payment Gateway (Bank BCA, Mandiri, BNI, QRIS).
  - Menampilkan perhitungan signature **HMAC-SHA256** secara transparan (`secret + raw_payload`).
  - Fitur pengujian skenario:
    1. **Bayar Normal (Success)**: Menghasilkan webhook valid dan status seketika menjadi `paid`.
    2. **Simulasi Idempoten (Duplicate)**: Mengirimkan callback kedua kali untuk membuktikan sistem tidak double-credit.
    3. **Simulasi Tampering**: Merusak signature untuk memvalidasi proteksi `401 Unauthorized`.
    4. **Simulasi Expired**: Menguji transisi invoice kadaluwarsa.

---

## 4. State Management & API Client

### 4.1 Axios Interceptor (Auto-refresh JWT)
```typescript
// src/core/http/apiClient.ts
import axios from 'axios';
import { useAuthStore } from '@/application/auth/useAuthStore';

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      try {
        const refreshToken = useAuthStore.getState().refreshToken;
        const res = await axios.post(`${apiClient.defaults.baseURL}/auth/refresh`, {
          refresh_token: refreshToken,
        });
        const { access_token } = res.data.data;
        useAuthStore.getState().setAccessToken(access_token);
        originalRequest.headers.Authorization = `Bearer ${access_token}`;
        return apiClient(originalRequest);
      } catch (refreshErr) {
        useAuthStore.getState().logout();
        window.location.href = '/login';
        return Promise.reject(refreshErr);
      }
    }
    return Promise.reject(error);
  }
);
```

---

## 5. UI/UX & Design Guidelines

1. **Aesthetics**: Desain modern dan clean menggunakan komponen Tailwind CSS + Radix UI dengan palet warna enterprise (Primary: Deep Blue `#1E40AF`, Success: Emerald `#10B981`, Warning/Unpaid: Amber `#F59E0B`, Danger: Rose `#EF4444`).
2. **Micro-Interactions**: Loading spinners, skeleton loaders saat fetching query TanStack, dan toast notifications (Sonner / Toast) untuk setiap aksi transaksi.
3. **Responsif**: Kompatibel dari smartphone 360px hingga layar desktop 1920px.
