import { describe, expect, test } from "vitest";

import { gitUrlSchema } from "../schema";

describe("Git url validation", () => {
  test("it validates a valid url", async () => {
    expect(
      gitUrlSchema.safeParse("git@examle.com:user/repository.git").success
    ).toBeTruthy();
  });

  test("it does not allow a http url", () => {
    expect(
      gitUrlSchema.safeParse("http://examle.com/user/repository.git").success
    ).toBeFalsy();
  });

  test("it does not allow a https url", () => {
    expect(
      gitUrlSchema.safeParse("https://examle.com/user/repository.git").success
    ).toBeFalsy();
  });

  test("it does not allow a slash after the domain name", () => {
    expect(
      gitUrlSchema.safeParse("git@examle.com/user/repository.git").success
    ).toBeFalsy();
  });

  test("it requires the url to start with git@", () => {
    expect(
      gitUrlSchema.safeParse("examle.com:user/repository.git").success
    ).toBeFalsy();
  });
});
