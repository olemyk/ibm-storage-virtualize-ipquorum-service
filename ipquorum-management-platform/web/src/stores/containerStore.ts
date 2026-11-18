import { create } from 'zustand';
import {
  ContainerInstance,
  CreateContainerRequest,
  ContainerFilters,
  ContainerSortOptions,
} from '../types/container';
import { containerApi } from '../services/containerApi';

interface ContainerState {
  // State
  containers: ContainerInstance[];
  selectedContainer: ContainerInstance | null;
  isLoading: boolean;
  error: string | null;
  filters: ContainerFilters;
  sortOptions: ContainerSortOptions;
  total: number;

  // Actions
  fetchContainers: () => Promise<void>;
  fetchContainer: (id: string) => Promise<void>;
  fetchContainerByInstanceId: (instanceId: string) => Promise<void>;
  createContainer: (data: CreateContainerRequest) => Promise<ContainerInstance>;
  startContainer: (id: string) => Promise<void>;
  stopContainer: (id: string, timeout?: number) => Promise<void>;
  restartContainer: (id: string, timeout?: number) => Promise<void>;
  deleteContainer: (id: string, force?: boolean) => Promise<void>;
  setSelectedContainer: (container: ContainerInstance | null) => void;
  setFilters: (filters: Partial<ContainerFilters>) => void;
  setSortOptions: (sortOptions: ContainerSortOptions) => void;
  clearError: () => void;
  reset: () => void;
}

const initialState = {
  containers: [],
  selectedContainer: null,
  isLoading: false,
  error: null,
  filters: {},
  sortOptions: {
    field: 'created_at' as const,
    order: 'desc' as const,
  },
  total: 0,
};

export const useContainerStore = create<ContainerState>((set, get) => ({
  ...initialState,

  fetchContainers: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await containerApi.listContainers();
      // Handle both direct array response and wrapped response (v2)
      const containers = Array.isArray(response) ? response : (response.containers || []);
      const total = Array.isArray(response) ? response.length : (response.total || 0);
      set({
        containers,
        total,
        isLoading: false,
        error: null,
      });
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error || error.message || 'Failed to fetch containers';
      set({
        containers: [],
        total: 0,
        isLoading: false,
        error: errorMessage,
      });
      throw error;
    }
  },

  fetchContainer: async (id: string) => {
    set({ isLoading: true, error: null });
    try {
      const container = await containerApi.getContainer(id);
      set({
        selectedContainer: container,
        isLoading: false,
        error: null,
      });
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error || error.message || 'Failed to fetch container';
      set({
        selectedContainer: null,
        isLoading: false,
        error: errorMessage,
      });
      throw error;
    }
  },

  fetchContainerByInstanceId: async (instanceId: string) => {
    set({ isLoading: true, error: null });
    try {
      const container = await containerApi.getContainerByInstanceId(instanceId);
      set({
        selectedContainer: container,
        isLoading: false,
        error: null,
      });
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error || error.message || 'Failed to fetch container';
      set({
        selectedContainer: null,
        isLoading: false,
        error: errorMessage,
      });
      throw error;
    }
  },

  createContainer: async (data: CreateContainerRequest) => {
    set({ isLoading: true, error: null });
    try {
      const container = await containerApi.createContainer(data);
      
      // Add to containers list
      set((state) => ({
        containers: [container, ...state.containers],
        total: state.total + 1,
        isLoading: false,
        error: null,
      }));
      
      return container;
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error || error.message || 'Failed to create container';
      set({
        isLoading: false,
        error: errorMessage,
      });
      throw error;
    }
  },

  startContainer: async (id: string) => {
    set({ isLoading: true, error: null });
    try {
      await containerApi.startContainer(id);
      
      // Update container status in list
      set((state) => ({
        containers: state.containers.map((c) =>
          c.id === id ? { ...c, status: 'running' as const } : c
        ),
        selectedContainer:
          state.selectedContainer?.id === id
            ? { ...state.selectedContainer, status: 'running' as const }
            : state.selectedContainer,
        isLoading: false,
        error: null,
      }));
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error || error.message || 'Failed to start container';
      set({
        isLoading: false,
        error: errorMessage,
      });
      throw error;
    }
  },

  stopContainer: async (id: string, timeout?: number) => {
    set({ isLoading: true, error: null });
    try {
      await containerApi.stopContainer(id, timeout);
      
      // Update container status in list
      set((state) => ({
        containers: state.containers.map((c) =>
          c.id === id ? { ...c, status: 'stopped' as const } : c
        ),
        selectedContainer:
          state.selectedContainer?.id === id
            ? { ...state.selectedContainer, status: 'stopped' as const }
            : state.selectedContainer,
        isLoading: false,
        error: null,
      }));
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error || error.message || 'Failed to stop container';
      set({
        isLoading: false,
        error: errorMessage,
      });
      throw error;
    }
  },

  restartContainer: async (id: string, timeout?: number) => {
    set({ isLoading: true, error: null });
    try {
      await containerApi.restartContainer(id, timeout);
      
      // Update container status in list
      set((state) => ({
        containers: state.containers.map((c) =>
          c.id === id ? { ...c, status: 'running' as const } : c
        ),
        selectedContainer:
          state.selectedContainer?.id === id
            ? { ...state.selectedContainer, status: 'running' as const }
            : state.selectedContainer,
        isLoading: false,
        error: null,
      }));
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error || error.message || 'Failed to restart container';
      set({
        isLoading: false,
        error: errorMessage,
      });
      throw error;
    }
  },

  deleteContainer: async (id: string, force: boolean = false) => {
    set({ isLoading: true, error: null });
    try {
      await containerApi.deleteContainer(id, force);
      
      // Remove from containers list
      set((state) => ({
        containers: state.containers.filter((c) => c.id !== id),
        total: state.total - 1,
        selectedContainer:
          state.selectedContainer?.id === id ? null : state.selectedContainer,
        isLoading: false,
        error: null,
      }));
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error || error.message || 'Failed to delete container';
      set({
        isLoading: false,
        error: errorMessage,
      });
      throw error;
    }
  },

  setSelectedContainer: (container: ContainerInstance | null) => {
    set({ selectedContainer: container });
  },

  setFilters: (filters: Partial<ContainerFilters>) => {
    set((state) => ({
      filters: { ...state.filters, ...filters },
    }));
  },

  setSortOptions: (sortOptions: ContainerSortOptions) => {
    set({ sortOptions });
  },

  clearError: () => {
    set({ error: null });
  },

  reset: () => {
    set(initialState);
  },
}));

// Selectors
export const selectFilteredContainers = (state: ContainerState) => {
  let filtered = [...state.containers];

  // Apply filters
  if (state.filters.status) {
    filtered = filtered.filter((c) => c.status === state.filters.status);
  }
  if (state.filters.runtime) {
    filtered = filtered.filter((c) => c.runtime === state.filters.runtime);
  }
  if (state.filters.agent_id) {
    filtered = filtered.filter((c) => c.agent_id === state.filters.agent_id);
  }
  if (state.filters.instance_id) {
    filtered = filtered.filter((c) => c.instance_id === state.filters.instance_id);
  }
  if (state.filters.search) {
    const search = state.filters.search.toLowerCase();
    filtered = filtered.filter(
      (c) =>
        c.container_name.toLowerCase().includes(search) ||
        c.image_name.toLowerCase().includes(search) ||
        c.id.toLowerCase().includes(search)
    );
  }

  // Apply sorting
  filtered.sort((a, b) => {
    const field = state.sortOptions.field;
    const order = state.sortOptions.order;
    
    let aVal: any = a[field];
    let bVal: any = b[field];
    
    // Handle date fields
    if (field === 'created_at' || field === 'updated_at') {
      aVal = new Date(aVal).getTime();
      bVal = new Date(bVal).getTime();
    }
    
    // Handle string fields
    if (field === 'container_name' || field === 'status') {
      aVal = String(aVal).toLowerCase();
      bVal = String(bVal).toLowerCase();
    }
    
    if (order === 'asc') {
      return aVal > bVal ? 1 : -1;
    } else {
      return aVal < bVal ? 1 : -1;
    }
  });

  return filtered;
};

export const selectContainerById = (state: ContainerState, id: string) => {
  return state.containers.find((c) => c.id === id);
};

export const selectContainersByStatus = (state: ContainerState, status: string) => {
  return state.containers.filter((c) => c.status === status);
};

export const selectContainerStats = (state: ContainerState) => {
  const stats = {
    total: state.containers.length,
    running: 0,
    stopped: 0,
    error: 0,
    created: 0,
    paused: 0,
  };

  state.containers.forEach((c) => {
    if (c.status in stats) {
      stats[c.status as keyof typeof stats]++;
    }
  });

  return stats;
};

// Made with Bob
