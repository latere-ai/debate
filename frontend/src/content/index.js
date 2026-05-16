import { computed } from 'vue';
import { usePrefsStore } from '../stores/prefs';
import { en } from './en';
import { zh } from './zh';
const dicts = { en, zh };
/**
 * Reactive landing content for the active locale. Consumers read
 * `content.value.hero.title` so a locale switch re-renders in place.
 */
export function useContent() {
    const prefs = usePrefsStore();
    return computed(() => dicts[prefs.locale] ?? en);
}
