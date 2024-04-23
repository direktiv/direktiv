import { beforeEach, describe, expect, test } from "vitest";
import {
  getTimeByIncrementAndType,
  getTimeFromDate,
  updateDateByTime,
} from "../utils";

describe("getTimeFromDate", () => {
  const date = new Date("2024-12-24T03:24:11");

  test("should extract the hours from a date value", () => {
    expect(getTimeFromDate(date, "hours")).toEqual("03");
  });

  test("should extract the minutes from a date value", () => {
    expect(getTimeFromDate(date, "minutes")).toEqual("24");
  });

  test("should extract the seconds from a date value", () => {
    expect(getTimeFromDate(date, "seconds")).toEqual("11");
  });
});

describe("getTimeByIncrementAndType", () => {
  test("should increment hours", () => {
    expect(getTimeByIncrementAndType("04", +1, "hours")).toEqual("05");
  });

  test("should increment hours and wrap around from 23 to 00", () => {
    expect(getTimeByIncrementAndType("23", +1, "hours")).toEqual("00");
  });

  test("should increment minutes", () => {
    expect(getTimeByIncrementAndType("27", +1, "minutes")).toEqual("28");
  });

  test("should increment minutes and wrap around from 59 to 00", () => {
    expect(getTimeByIncrementAndType("59", +1, "minutes")).toEqual("00");
  });

  test("should increment seconds", () => {
    expect(getTimeByIncrementAndType("45", +1, "seconds")).toEqual("46");
  });

  test("should increment seconds and wrap around from 59 to 00", () => {
    expect(getTimeByIncrementAndType("59", +1, "seconds")).toEqual("00");
  });

  test("should decrement hours", () => {
    expect(getTimeByIncrementAndType("05", -1, "hours")).toEqual("04");
  });

  test("should decrement hours and wrap around from 00 to 23", () => {
    expect(getTimeByIncrementAndType("00", -1, "hours")).toEqual("23");
  });

  test("should decrement minutes", () => {
    expect(getTimeByIncrementAndType("28", -1, "minutes")).toEqual("27");
  });

  test("should decrement minutes and wrap around from 00 to 59", () => {
    expect(getTimeByIncrementAndType("00", -1, "minutes")).toEqual("59");
  });

  test("should decrement seconds", () => {
    expect(getTimeByIncrementAndType("46", -1, "seconds")).toEqual("45");
  });

  test("should decrement seconds and wrap around from 00 to 59", () => {
    expect(getTimeByIncrementAndType("00", -1, "seconds")).toEqual("59");
  });
});

describe("updateDateByTime", () => {
  let date: Date;

  beforeEach(() => {
    date = new Date("2024-04-09T08:37:21");
  });

  test("should update minutes", () => {
    const updatedDate = updateDateByTime(date, "45", "minutes");
    expect(updatedDate.getMinutes()).toEqual(45);
    expect(updatedDate.getSeconds()).toEqual(21);
    expect(updatedDate.getHours()).toEqual(8);
  });

  test("should update seconds", () => {
    const updatedDate = updateDateByTime(date, "15", "seconds");
    expect(updatedDate.getSeconds()).toEqual(15);
    expect(updatedDate.getMinutes()).toEqual(37);
    expect(updatedDate.getHours()).toEqual(8);
  });

  test("should update hours", () => {
    const updatedDate = updateDateByTime(date, "18", "hours");
    expect(updatedDate.getHours()).toEqual(18);
    expect(updatedDate.getMinutes()).toEqual(37);
    expect(updatedDate.getSeconds()).toEqual(21);
  });
});
