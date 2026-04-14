import { describe, it, expect } from "vitest";

describe("Smoke test", () => {
  it("types are importable", async () => {
    const types = await import("@/types");
    expect(types).toBeDefined();
  });

  it("api client is importable", async () => {
    const { api } = await import("@/lib/api-client");
    expect(api).toBeDefined();
    expect(api.get).toBeTypeOf("function");
    expect(api.post).toBeTypeOf("function");
  });
});
