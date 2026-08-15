import { create } from 'zustand';

/**
 * Small client-only state.
 *
 * Server data belongs in TanStack Query; this store holds nothing that the API
 * owns (docs/06-mobile-architecture.md).
 */
interface UIState {
  /** The senior currently in focus. A caregiver may manage several. */
  selectedSeniorId: string | null;
  setSelectedSeniorId: (seniorId: string | null) => void;

  /** Whether the user has completed the onboarding flow on this device. */
  hasCompletedOnboarding: boolean;
  setHasCompletedOnboarding: (completed: boolean) => void;

  reset: () => void;
}

export const useUIStore = create<UIState>((set) => ({
  selectedSeniorId: null,
  setSelectedSeniorId: (selectedSeniorId) => set({ selectedSeniorId }),

  hasCompletedOnboarding: false,
  setHasCompletedOnboarding: (hasCompletedOnboarding) => set({ hasCompletedOnboarding }),

  reset: () => set({ selectedSeniorId: null, hasCompletedOnboarding: false }),
}));
