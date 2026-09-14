import { expect, test } from "@playwright/test";

test.describe("Login page", () => {
  test("renders the login form and account links", async ({ page }) => {
    await page.goto("/login");

    await expect(page.getByRole("heading", { name: "Welcome back" })).toBeVisible();
    await expect(page.getByLabel("Email address")).toBeVisible();
    await expect(page.locator("#password")).toBeVisible();
    await expect(page.getByRole("button", { name: "Sign in" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Create an account" })).toHaveAttribute(
      "href",
      "/register",
    );
    await expect(page.getByRole("link", { name: "Forgot password" })).toHaveAttribute(
      "href",
      "/forgot-password",
    );
  });

  test("shows validation errors without submitting invalid form data", async ({ page }) => {
    await page.goto("/login");

    await page.getByRole("button", { name: "Sign in" }).click();

    await expect(page.locator("#email-error")).toHaveText("Email is required");
    await expect(page.locator("#password-error")).toHaveText("Password is required");

    await page.getByLabel("Email address").fill("not-an-email");
    await page.locator("#password").fill("short");
    await page.getByRole("button", { name: "Sign in" }).click();

    await expect(page.locator("#email-error")).toHaveText("Please enter a valid email address");
    await expect(page.locator("#password-error")).toHaveText(
      "Password must be at least 8 characters long",
    );
  });

  test("shows the backend error for invalid credentials", async ({ page }) => {
    const loginRequest = page.waitForRequest(
      (request) => request.url().endsWith("/api/login") && request.method() === "POST",
    );

    await page.route("**/api/login", async (route) => {
      await route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({ detail: "Invalid email or password" }),
      });
    });

    await page.goto("/login");
    await page.getByLabel("Email address").fill("user@example.com");
    await page.locator("#password").fill("wrongpass");
    await page.getByRole("button", { name: "Sign in" }).click();

    const request = await loginRequest;
    expect(request.postDataJSON()).toEqual({
      email: "user@example.com",
      password: "wrongpass",
    });
    await expect(page.locator('p[role="alert"]')).toHaveText("Invalid email or password");
    await expect(page.getByRole("button", { name: "Sign in" })).toBeEnabled();
  });

  test("redirects to the requested URL after a successful login", async ({ page }) => {
    await page.route("**/api/login", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ message: "Login successful" }),
      });
    });

    await page.goto("/login?redirect=%2Flogin%3Fregistered%3Dtrue");
    await page.getByLabel("Email address").fill("user@example.com");
    await page.locator("#password").fill("correctpass");

    await Promise.all([
      page.waitForURL("**/login?registered=true"),
      page.getByRole("button", { name: "Sign in" }).click(),
    ]);

    await expect(page.getByRole("status")).toHaveText("Your account is ready. Sign in to get started.");
  });
});
