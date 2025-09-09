import { BlockType } from "../../../schema/blocks";
import { DirektivPagesType } from "../../../schema";
import { MutationType } from "../../../schema/procedures/mutation";

// minimal ResizeObserver mock required by radix-ui checkbox
// https://github.com/radix-ui/primitives/blob/main/packages/react/checkbox/src/checkbox.test.tsx#L11
export const setupResizeObserverMock = () => {
  global.ResizeObserver = class ResizeObserver {
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    observe() {}
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    unobserve() {}
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    disconnect() {}
  };
};

export const createDirektivPage = (blocks: BlockType[]): DirektivPagesType => ({
  direktiv_api: "page/v1",
  type: "page",
  blocks,
});

export const createDirektivPageWithForm = (
  blocks: BlockType[],
  mutation: MutationType = {
    method: "POST",
    url: "/some-endpoint",
  }
) =>
  createDirektivPage([
    {
      type: "query-provider",
      queries: [
        {
          id: "user",
          url: "/user-details",
          queryParams: [],
        },
      ],
      blocks: [
        {
          type: "form",
          trigger: {
            type: "button",
            label: "save",
          },
          mutation,
          blocks,
        },
      ],
    },
  ]);

export const setPage = (page: DirektivPagesType) => page;
