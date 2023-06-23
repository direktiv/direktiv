import { Locator } from "@playwright/test";

export const getStyle = async (
  locator: Locator,
  property: string
): Promise<string> =>
  locator.evaluate(
    (el, property) => window.getComputedStyle(el).getPropertyValue(property),
    property
  );
