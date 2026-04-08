import { describe, expect, it } from "vitest";
import { generateElbowPathParams, generateElbowPaths } from "./paths";

// these same values are used as the expected output of
// generateElbowPathParams and the input for generateElbowPaths
const pathParams = [
  {
    endY: 98,
    s: -1,
    targetY: 49,
    ya: 282,
    yb: 61,
  },
  {
    endY: 392,
    s: -1,
    targetY: 245,
    ya: 282,
    yb: 257,
  },
  {
    endY: 490,
    s: 1,
    targetY: 441,
    ya: 306,
    yb: 429,
  },
  {
    endY: 686,
    s: 1,
    targetY: 588,
    ya: 306,
    yb: 576,
  },
];

// the above results in these svg paths
const expectedPaths = [
  <path
    key="0"
    d="M0,294
H 20
A 12 12 0 0 0 32 282
V 61
A 12 12 0 0 1 44 49
H 64"
  />,
  <path
    key="1"
    d="M0,294
H 20
A 12 12 0 0 0 32 282
V 257
A 12 12 0 0 1 44 245
H 64"
  />,
  <path
    key="2"
    d="M0,294
H 20
A 12 12 0 0 1 32 306
V 429
A 12 12 0 0 0 44 441
H 64"
  />,
  <path
    key="3"
    d="M0,294
H 20
A 12 12 0 0 1 32 306
V 576
A 12 12 0 0 0 44 588
H 64"
  />,
];

// with reverse=true, it will result in these svg paths
const expectedPathsReversed = [
  <path
    key="0"
    d="M0,294
H 20
A 12 12 0 0 1 32 282
V 61
A 12 12 0 0 0 44 49
H 64"
  />,
  <path
    key="1"
    d="M0,294
H 20
A 12 12 0 0 1 32 282
V 257
A 12 12 0 0 0 44 245
H 64"
  />,
  <path
    key="2"
    d="M0,294
H 20
A 12 12 0 0 0 32 306
V 429
A 12 12 0 0 1 44 441
H 64"
  />,
  <path
    key="3"
    d="M0,294
H 20
A 12 12 0 0 0 32 306
V 576
A 12 12 0 0 1 44 588
H 64"
  />,
];

// another test case containing a simple horizontal line
const expectedPathsWithLine = [
  <path
    key="0"
    d="M0,144
H 20
A 12 12 0 0 0 32 132
V 60
A 12 12 0 0 1 44 48
H 64"
  />,
  <line key="1" x1={0} y1={144} x2={64} y2={144} />,
  <path
    key="2"
    d="M0,144
H 20
A 12 12 0 0 1 32 156
V 228
A 12 12 0 0 0 44 240
H 64"
  ></path>,
];

describe("generateElbowPathParams", () => {
  it("should calculate path params", () => {
    const result = generateElbowPathParams([1, 3, 1, 2], 294, 98, 12);
    expect(result).toStrictEqual(pathParams);
  });
});

describe("generateElbowPaths", () => {
  it("should calculate path params for an opening elbow bracket", () => {
    const context = {
      startX: 0,
      startY: 294,
      endX: 64,
      xl: 20,
      xm: 32,
      xr: 44,
      r: 12,
      reverse: false,
    };

    const result = generateElbowPaths(pathParams, context);

    // if this fails, vitest may not report the differences.
    // using JSON.stringify(value) for both values may help in this case.
    expect(result).toEqual(expectedPaths);
  });

  it("should calculate path params for a closing elbow bracket", () => {
    const context = {
      startX: 0,
      startY: 294,
      endX: 64,
      xl: 20,
      xm: 32,
      xr: 44,
      r: 12,
      reverse: true,
    };

    const result = generateElbowPaths(pathParams, context);

    // if this fails, vitest may not report the differences.
    // using JSON.stringify(value) for both values may help in this case.
    expect(result).toEqual(expectedPathsReversed);
  });

  it("should draw a straight line if one segment starts and ends on the same y coordinate", () => {
    const context = {
      startX: 0,
      startY: 144,
      endX: 64,
      xl: 20,
      xm: 32,
      xr: 44,
      r: 12,
      reverse: false,
    };

    const pathParams = [
      { s: -1, endY: 96, targetY: 48, ya: 132, yb: 60 },
      { s: -1, endY: 192, targetY: 144, ya: 132, yb: 156 },
      { s: 1, endY: 288, targetY: 240, ya: 156, yb: 228 },
    ];

    const result = generateElbowPaths(pathParams, context);

    // if this fails, vitest may not report the differences.
    // using JSON.stringify(value) for both values may help in this case.
    expect(result).toEqual(expectedPathsWithLine);
  });
});
