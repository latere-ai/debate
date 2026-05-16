import { describe, it, expect } from 'vitest';
import { en } from './en';
import { zh } from './zh';
/** Collect leaf paths + whether each leaf is an empty string. */
function walk(node, path, out) {
    if (Array.isArray(node)) {
        out.push({ path: `${path}[]`, empty: false }); // record length via marker below
        node.forEach((v, i) => walk(v, `${path}[${i}]`, out));
    }
    else if (node !== null && typeof node === 'object') {
        for (const k of Object.keys(node).sort()) {
            walk(node[k], path ? `${path}.${k}` : k, out);
        }
    }
    else {
        out.push({ path, empty: typeof node === 'string' && node.trim() === '' });
    }
}
describe('landing content dictionaries', () => {
    const enLeaves = [];
    const zhLeaves = [];
    walk(en, '', enLeaves);
    walk(zh, '', zhLeaves);
    it('en and zh have identical structure (keys + array lengths)', () => {
        const enPaths = enLeaves.map(l => l.path);
        const zhPaths = zhLeaves.map(l => l.path);
        const missingInZh = enPaths.filter(p => !zhPaths.includes(p));
        const extraInZh = zhPaths.filter(p => !enPaths.includes(p));
        expect(missingInZh, `zh missing: ${missingInZh.join(', ')}`).toEqual([]);
        expect(extraInZh, `zh extra: ${extraInZh.join(', ')}`).toEqual([]);
    });
    it('a value is empty in zh only where it is empty in en (catches missing translations; allows the intentional terminal-diagram spacer)', () => {
        const zhByPath = new Map(zhLeaves.map(l => [l.path, l.empty]));
        const mismatched = enLeaves
            .filter(l => zhByPath.has(l.path) && zhByPath.get(l.path) !== l.empty)
            .map(l => l.path);
        expect(mismatched, `emptiness differs en↔zh at: ${mismatched.join(', ')}`).toEqual([]);
    });
});
