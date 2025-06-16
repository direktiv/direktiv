import { variablePattern } from "../Variable/utils";

// TODO: add tests
export const processTemplateString = <T>(
  value: string,
  onMatch: (match: string, index: number) => T
): (string | T)[] => {
  const fragments = value.split(variablePattern);

  return fragments.map((fragment, index) => {
    const isVariable = index % 2 === 1;

    if (isVariable) {
      return onMatch(fragment, index);
    }

    return fragment;
  });
};
