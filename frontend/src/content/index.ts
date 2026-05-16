import { computed, type ComputedRef } from 'vue';
import { usePrefsStore } from '../stores/prefs';
import type { LandingContent } from './types';
import { en } from './en';
import { zh } from './zh';

export type { LandingContent } from './types';

const dicts: Record<string, LandingContent> = { en, zh };

/**
 * Reactive landing content for the active locale. Consumers read
 * `content.value.hero.title` so a locale switch re-renders in place.
 */
export function useContent(): ComputedRef<LandingContent> {
  const prefs = usePrefsStore();
  return computed(() => dicts[prefs.locale] ?? en);
}
