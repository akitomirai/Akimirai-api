import { describe, expect, it } from 'vitest'

import { findRowIndexByDomPosition } from '../useSwipeSelect'

describe('findRowIndexByDomPosition', () => {
  it('locates fully rendered rows and clamps outside coordinates', () => {
    const table = document.createElement('table')
    const tbody = document.createElement('tbody')
    table.appendChild(tbody)

    for (let index = 0; index < 3; index += 1) {
      const row = document.createElement('tr')
      row.dataset.index = String(index + 10)
      row.getBoundingClientRect = () => ({
        top: index * 40,
        bottom: index * 40 + 30,
        left: 0,
        right: 100,
        width: 100,
        height: 30,
        x: 0,
        y: index * 40,
        toJSON: () => ({})
      })
      tbody.appendChild(row)
    }

    expect(findRowIndexByDomPosition(table, -5)).toBe(10)
    expect(findRowIndexByDomPosition(table, 45)).toBe(11)
    expect(findRowIndexByDomPosition(table, 75)).toBe(12)
    expect(findRowIndexByDomPosition(table, 200)).toBe(12)
  })
})
