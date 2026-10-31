import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/core/http/apiClient';
import { ApiResponse, Invoice, Institution, KrsStatus, LoginResponse, TransactionHistory } from '@/domain/models/types';
import { useAuthStore } from '@/application/auth/useAuthStore';

export const QUERY_KEYS = {
  invoices: ['invoices'],
  invoiceDetail: (id: string) => ['invoice', id],
  krsStatus: (sem: string, yr: string) => ['krs-status', sem, yr],
  institutions: ['institutions'],
};

// ---------------- Auth Queries & Mutations ----------------

export function useLoginMutation() {
  const setAuth = useAuthStore((s) => s.setAuth);
  return useMutation({
    mutationFn: async (credentials: { email: string; password: string }) => {
      const res = await apiClient.post<ApiResponse<LoginResponse>>('/auth/login', credentials);
      return res.data.data;
    },
    onSuccess: (data) => {
      setAuth({
        user: data.user,
        accessToken: data.access_token,
        refreshToken: data.refresh_token,
        institutionCode: data.institution_code,
        institutionName: data.institution_name,
      });
    },
  });
}

// ---------------- Invoices Queries & Mutations ----------------

export interface InvoiceFilterParams {
  status?: string;
  semester?: string;
  academic_year?: string;
  search?: string;
  page?: number;
  per_page?: number;
}

export function useInvoicesQuery(params: InvoiceFilterParams = {}) {
  return useQuery({
    queryKey: [...QUERY_KEYS.invoices, params],
    queryFn: async () => {
      const res = await apiClient.get<ApiResponse<Invoice[]>>('/invoices', { params });
      return res.data;
    },
  });
}

export function useInvoiceDetailQuery(id: string) {
  return useQuery({
    queryKey: QUERY_KEYS.invoiceDetail(id),
    queryFn: async () => {
      const res = await apiClient.get<ApiResponse<{ invoice: Invoice; history: TransactionHistory[] }>>(`/invoices/${id}`);
      return res.data.data;
    },
    enabled: Boolean(id),
  });
}

export function useCreateInvoiceMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (payload: {
      student_id: string;
      amount: number;
      semester: string;
      academic_year: string;
      due_date: string;
      notes?: string;
    }) => {
      const res = await apiClient.post<ApiResponse<Invoice>>('/invoices', payload);
      return res.data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.invoices });
    },
  });
}

// ---------------- KRS Status Query ----------------

export function useKrsStatusQuery(semester = 'ganjil', academicYear = '2026/2027') {
  return useQuery({
    queryKey: QUERY_KEYS.krsStatus(semester, academicYear),
    queryFn: async () => {
      const res = await apiClient.get<ApiResponse<KrsStatus>>('/students/krs-status', {
        params: { semester, academic_year: academicYear },
      });
      return res.data.data;
    },
  });
}

// ---------------- Institutions Queries ----------------

export function useInstitutionsQuery() {
  return useQuery({
    queryKey: QUERY_KEYS.institutions,
    queryFn: async () => {
      const res = await apiClient.get<ApiResponse<Institution[]>>('/institutions');
      return res.data.data;
    },
  });
}

export function useCreateInstitutionMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (payload: {
      code: string;
      name: string;
      email: string;
      phone?: string;
      address?: string;
    }) => {
      const res = await apiClient.post<ApiResponse<Institution>>('/institutions', payload);
      return res.data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.institutions });
    },
  });
}

// ---------------- Mock Payment Gateway Mutation ----------------

export interface MockPaymentPayload {
  institution_code: string;
  invoice_number: string;
  payment_method: string;
  simulate_tamper?: boolean;
  event_type?: string;
}

export function useMockPaymentMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (payload: MockPaymentPayload) => {
      const res = await apiClient.post<ApiResponse<any>>('/mock-gateway/pay', payload);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.invoices });
      queryClient.invalidateQueries({ queryKey: ['krs-status'] });
    },
  });
}
