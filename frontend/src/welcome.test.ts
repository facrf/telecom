import { expect, test } from 'vitest'
import { welcomeTitle } from './welcome'

test('exposes the product title', () => {
  expect(welcomeTitle).toContain('Telecomunicações')
})
