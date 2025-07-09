import { add, sub } from '../src/math.js';

describe('add', () => {
  it('adds two positive integers', () => {
    expect(add(2, 3)).toBe(5);
  });

  it('adds positive and negative numbers', () => {
    expect(add(5, -2)).toBe(3);
  });

  it('handles zero edge-case', () => {
    expect(add(0, 4)).toBe(4);
  });

  it('throws TypeError for non-number inputs', () => {
    expect(() => add('a', 1)).toThrow(TypeError);
  });
});

describe('sub', () => {
  it('subtracts two positive integers', () => {
    expect(sub(5, 2)).toBe(3);
  });

  it('subtracts positive and negative numbers', () => {
    expect(sub(5, -2)).toBe(7);
  });

  it('handles zero edge-case', () => {
    expect(sub(0, 5)).toBe(-5);
  });

  it('throws TypeError for non-number inputs', () => {
    expect(() => sub(1, 'b')).toThrow(TypeError);
  });
});
