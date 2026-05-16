import { describe, it, expect } from 'vitest';
const fs = (await import('node:fs'));
const path = (await import('node:path'));
const css = fs.readFileSync(path.resolve(process.cwd(), 'src/styles/dialectic.css'), 'utf8');
// Regression guard for the dark-mode spine-stop bug: the generic
// `[data-theme="dark"] .v-dialectic .spine-stop { background: var(--bg-raised) }`
// rule has the same specificity as the base `.spine-stop.is-c` rule and
// comes later, so it overrode is-c/is-stake backgrounds while their light
// text colour stayed — dark-on-dark, invisible. The fix is an explicit
// dark is-c/is-stake override placed AFTER the generic dark rule.
describe('dialectic.css dark spine-stop', () => {
    const genericDark = css.indexOf('[data-theme="dark"] .v-dialectic .spine-stop {');
    const isCDark = css.indexOf('[data-theme="dark"] .v-dialectic .spine-stop.is-c');
    it('has an explicit dark is-c / is-stake override', () => {
        expect(genericDark, 'generic dark .spine-stop rule missing').toBeGreaterThan(-1);
        expect(isCDark, 'dark .spine-stop.is-c override missing').toBeGreaterThan(-1);
    });
    it('places the is-c override AFTER the generic dark rule (cascade order)', () => {
        expect(isCDark).toBeGreaterThan(genericDark);
    });
    it('restores a filled crimson stop with readable text', () => {
        const block = css.slice(isCDark, isCDark + 240);
        expect(block).toContain('.spine-stop.is-stake');
        expect(block).toMatch(/background:\s*var\(--accent\)/);
        expect(block).toMatch(/color:\s*var\(--bg\)/);
    });
});
