import { Tree, getSublist } from "../utils";
import { describe, expect, it } from "vitest";

describe("getSublist", () => {
  const sampleTree: Tree = {
    animals: {
      mammals: {
        dog: { breeds: {} },
        cat: {},
      },
      reptiles: {
        snake: {},
      },
    },
    plants: {
      flowers: {},
      trees: {},
    },
    misc: "not a tree",
  };

  it("returns top-level keys when path is empty", () => {
    expect(getSublist(sampleTree, [])?.sort()).toEqual(
      ["animals", "plants", "misc"].sort()
    );
  });

  it("returns keys of a valid subtree", () => {
    expect(getSublist(sampleTree, ["animals"])).toEqual([
      "mammals",
      "reptiles",
    ]);
  });

  it("returns keys of a deeper subtree", () => {
    expect(getSublist(sampleTree, ["animals", "mammals"])).toEqual([
      "dog",
      "cat",
    ]);
  });

  it("returns null if path points to a non-tree value", () => {
    expect(getSublist(sampleTree, ["misc"])).toBeNull();
  });

  it("returns null if path points to a missing branch", () => {
    expect(getSublist(sampleTree, ["nonexistent"])).toBeNull();
    expect(getSublist(sampleTree, ["animals", "fish"])).toBeNull();
  });

  it("handles an empty tree", () => {
    expect(getSublist({}, [])).toEqual([]);
  });

  it("returns null when navigating too deep into empty subtree", () => {
    expect(getSublist(sampleTree, ["plants", "flowers", "rose"])).toBeNull();
  });
});
