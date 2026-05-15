import { describe, it, expect } from 'vitest';
import { en } from './en';
import { zh } from './zh';

describe('i18n dictionaries', () => {
  const enKeys = Object.keys(en).sort();
  const zhKeys = Object.keys(zh).sort();

  it('zh covers every en key', () => {
    const missing = enKeys.filter(k => !(k in zh));
    expect(missing, `zh missing: ${missing.join(', ')}`).toEqual([]);
  });

  it('en covers every zh key', () => {
    const extra = zhKeys.filter(k => !(k in en));
    expect(extra, `zh extra: ${extra.join(', ')}`).toEqual([]);
  });

  it('no empty values', () => {
    expect(enKeys.filter(k => en[k].trim() === '')).toEqual([]);
    expect(zhKeys.filter(k => zh[k].trim() === '')).toEqual([]);
  });
});
