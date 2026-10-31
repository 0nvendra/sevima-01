export type UserRole = 'super_admin' | 'admin_keuangan' | 'mahasiswa';
export type InvoiceStatus = 'unpaid' | 'paid' | 'expired';
export type SemesterType = 'ganjil' | 'genap';

export interface User {
  id: string;
  institusi_id?: string;
  email: string;
  full_name: string;
  role: UserRole;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Institution {
  id: string;
  code: string;
  name: string;
  address?: string;
  phone?: string;
  email?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Student {
  id: string;
  user_id: string;
  institusi_id: string;
  nim: string;
  study_program: string;
  enrollment_year: number;
  is_active: boolean;
  user?: User;
}

export interface Invoice {
  id: string;
  institusi_id: string;
  student_id: string;
  invoice_number: string;
  amount: number;
  status: InvoiceStatus;
  semester: SemesterType;
  academic_year: string;
  due_date: string;
  paid_at?: string;
  paid_amount?: number;
  payment_method?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
  student?: Student;
}

export interface TransactionHistory {
  id: string;
  invoice_id: string;
  institusi_id: string;
  previous_status: InvoiceStatus;
  new_status: InvoiceStatus;
  source: string;
  notes?: string;
  created_at: string;
}

export interface KrsStatus {
  nim: string;
  full_name: string;
  is_eligible_krs: boolean;
  unpaid_invoices_count: number;
  active_semester: string;
  active_academic_year: string;
  message: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
  institution_code?: string;
  institution_name?: string;
}

export interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data: T;
  meta?: {
    page: number;
    per_page: number;
    total_records: number;
    total_pages: number;
  };
  error?: {
    code: string;
    message: string;
    details?: any;
  };
}
