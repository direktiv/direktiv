import { AndGroup, Connector, OrGroup } from "..";
import {
  type ClauseInput,
  type NodeVM,
  shouldRenderConnector,
  toAndBranch,
} from "./utils";
import {
  type ConditionsType,
  ConditionsSchema,
} from "~/pages/namespace/Explorer/policy/schema/primitives/conditions";

import { type Meta, type StoryObj } from "@storybook/react-vite";

import { Card } from "~/design/Card";
const meta = {
  title: "POC/Policy/Condition Flow",
  component: AndGroup,
} satisfies Meta<typeof AndGroup>;

export default meta;

type Story = StoryObj<typeof meta>;

const ExpressionLeaf = ({ preview }: { preview: string }) => (
  <Card
    background="weight-2"
    className="my-[16px] flex h-[64px] min-w-40 max-w-64 overflow-hidden px-3 py-2"
  >
    <pre
      title={preview}
      className="m-0 block max-h-full w-full overflow-hidden whitespace-pre-wrap break-words text-xs leading-tight"
    >
      {preview}
    </pre>
  </Card>
);

const RenderNode = ({ node }: { node: NodeVM }) => {
  if (node.type === "leaf") {
    return <ExpressionLeaf preview={node.preview} />;
  }

  if (node.type === "or") {
    return (
      <OrGroup childSizes={node.childSizes}>
        {node.branches.map((branch, index) => (
          <RenderNode key={index} node={branch} />
        ))}
      </OrGroup>
    );
  }

  return (
    <AndGroup>
      {node.items.flatMap((item, index) => {
        const nextItem = node.items[index + 1];
        const renderedItems = [
          <RenderNode key={`item-${index}`} node={item} />,
        ];

        if (nextItem !== undefined && shouldRenderConnector(item, nextItem)) {
          renderedItems.push(<Connector key={`connector-${index}`} />);
        }

        return renderedItems;
      })}
    </AndGroup>
  );
};

const ClauseBlock = ({ clause }: { clause: ClauseInput }) => (
  <div className="flex flex-col gap-3">
    <div className="flex items-center gap-3">
      <span className="rounded-full bg-gray-4 px-3 py-1 text-xs font-medium uppercase tracking-[0.16em] text-gray-11 dark:bg-gray-dark-4 dark:text-gray-dark-11">
        {clause.kind}
      </span>
    </div>
    <div className="overflow-x-auto rounded-lg border border-dashed border-gray-5 bg-gray-1/40 p-4 dark:border-gray-dark-5 dark:bg-gray-dark-2/40">
      <RenderNode node={toAndBranch(clause.body)} />
    </div>
  </div>
);

const PolicyConditionFlow = ({ clauses }: { clauses: ClauseInput[] }) => (
  <div className="flex max-w-full flex-col gap-8 p-8">
    {clauses.map((clause, index) => (
      <ClauseBlock key={`${clause.kind}-${index}`} clause={clause} />
    ))}
  </div>
);

const simpleWhen = [
  {
    kind: "when",
    body: {
      "&&": {
        left: { has: { left: { Var: "principal" }, attr: "email" } },
        right: {
          like: {
            left: { ".": { left: { Var: "principal" }, attr: "email" } },
            pattern: ["Wildcard", { Literal: "@example.com" }],
          },
        },
      },
    },
  },
] satisfies ConditionsType;

const nestedBooleanGroups = [
  {
    kind: "when",
    body: {
      "&&": {
        left: { has: { left: { Var: "principal" }, attr: "email" } },
        right: {
          "||": {
            left: {
              "&&": {
                left: {
                  in: {
                    left: { Var: "resource" },
                    right: {
                      Value: { __entity: { type: "Folder", id: "Public" } },
                    },
                  },
                },
                right: { Value: true },
              },
            },
            right: {
              "&&": {
                left: {
                  "==": {
                    left: { Var: "action" },
                    right: {
                      Value: { __entity: { type: "Action", id: "readFile" } },
                    },
                  },
                },
                right: {
                  "==": {
                    left: { Var: "context" },
                    right: { Value: { region: "eu-west-1" } },
                  },
                },
              },
            },
          },
        },
      },
    },
  },
] satisfies ConditionsType;

const whenAndUnless = [
  ...simpleWhen,
  {
    kind: "unless",
    body: {
      "||": {
        left: {
          getTag: {
            left: { Var: "resource" },
            right: { Value: "classification" },
          },
        },
        right: {
          "&&": {
            left: { Value: false },
            right: {
              is: {
                left: { Var: "principal" },
                entity_type: "User",
              },
            },
          },
        },
      },
    },
  },
] satisfies ConditionsType;

const renderFlow = (clauses: ConditionsType) => (
  <PolicyConditionFlow clauses={ConditionsSchema.parse(clauses)} />
);

export const SimpleWhen: Story = {
  render: () => renderFlow(simpleWhen),
};

export const NestedBooleanGroups: Story = {
  render: () => renderFlow(nestedBooleanGroups),
};

export const WhenAndUnless: Story = {
  render: () => renderFlow(whenAndUnless),
};
