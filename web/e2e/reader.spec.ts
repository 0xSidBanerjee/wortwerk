import { expect, test } from '@playwright/test'

test.describe('reader fixture', () => {
  test.use({ viewport: { width: 390, height: 844 } })

  test('tap word, popup, show more, resume after reload', async ({ page }) => {
    await page.goto('/s/fixture')
    await expect(page.getByRole('heading', { name: 'Fixture' })).toBeVisible()

    const steht = page.locator('[data-s="s0"][data-t="1"]')
    await steht.click()
    await expect(page.getByRole('dialog')).toBeVisible()
    await expect(page.getByText('to get up')).toBeVisible()

    await page.getByText('Show more').click()
    await expect(page.getByText('Pinocchio gets up early.')).toBeVisible()

    await page.getByRole('button', { name: 'Close' }).click()
    await expect(page.getByRole('dialog')).not.toBeVisible()

    await page.evaluate(() => {
      window.scrollTo(0, document.body.scrollHeight)
    })
    await page.waitForTimeout(600)

    await page.reload()
    await expect(page.getByRole('heading', { name: 'Fixture' })).toBeVisible()
  })
})
