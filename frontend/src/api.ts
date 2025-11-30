import axios, { type AxiosRequestConfig } from "axios";

// Type definitions
export interface DNS {
  name: string;
  ip: string;
  port: number;
  protocol: string;
}

export interface Host {
  id: string;
  name: string;
  protocol: string;
  hostUrl: string;
  port: number;
  pingInterval: number;
  failThreshold: number;
  acceptedStatusCodes: string[];
  status: string;
  dns?: DNS[];
  createdAt: string;
  updatedAt: string;
}

export interface Result {
  dns: string;
  statusCode: number;
  latency: string;
  timestamp: string;
  errorMsg?: string;
}

export interface HistoriesResponse {
  hostUrl: string;
  results: Result[];
}

export interface Notification {
  id: string;
  hostId: string;
  name: string;
  type: string;
  active: boolean;
  lastNotifiedAt: string;
  discordUsername?: string;
  discordWebhookUrl?: string;
  discordPrefixMessage?: string;
  discordDisableUrl?: boolean;
  discordChannelType?: string;
  discordThreadId?: string;
  discordPostName?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateHostRequest {
  name: string;
  protocol: string;
  hostUrl: string;
  port: number;
  pingInterval: number;
  failThreshold: number;
  acceptedStatusCodes: string[];
  dns?: DNS[];
}

export type UpdateHostRequest = Partial<CreateHostRequest>;

export interface CreateNotificationRequest {
  name: string;
  type: string;
  active: boolean;
  discordUsername?: string;
  discordWebhookUrl?: string;
  discordPrefixMessage?: string;
  discordDisableUrl?: boolean;
  discordChannelType?: string;
  discordThreadId?: string;
  discordPostName?: string;
}

export type UpdateNotificationRequest = Partial<CreateNotificationRequest>;

// New required signal type for GET requests
type GetRequestConfig = AxiosRequestConfig & { signal: AbortSignal };

const api = axios.create({
  baseURL: "/api/v1",
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const message =
      error.response?.data?.message ||
      error.message ||
      "An unknown error occurred";
    return Promise.reject(new Error(message));
  }
);

// Host APIs
export const getHosts = (config: GetRequestConfig) =>
  api.get<{ hosts: Host[] }>("/uptime/all", config);

export const createHost = (data: CreateHostRequest) =>
  api.post<Host>("/uptime/create", data);

export const updateHost = (id: string, data: UpdateHostRequest) =>
  api.put<Host>(`/uptime/${id}`, data);

export const deleteHost = (id: string) => api.delete(`/uptime/${id}`);

export const getHostHistory = (
  id: string,
  startAt: string,
  endAt: string,
  config: GetRequestConfig
) =>
  api.get(`/uptime/${id}/history`, { params: { startAt, endAt }, ...config });

// Notification APIs
export const getNotifications = (hostId: string, config: GetRequestConfig) =>
  api.get<Notification[]>(`/uptime/${hostId}/notification`, config);

export const createNotification = (
  hostId: string,
  data: CreateNotificationRequest
) => api.post<Notification>(`/uptime/${hostId}/notification`, data);

export const updateNotification = (
  hostId: string,
  notificationId: string,
  data: UpdateNotificationRequest
) =>
  api.put<Notification>(
    `/uptime/${hostId}/notification/${notificationId}`,
    data
  );

export const deleteNotification = (hostId: string, notificationId: string) =>
  api.delete(`/uptime/${hostId}/notification/${notificationId}`);

export default api;
