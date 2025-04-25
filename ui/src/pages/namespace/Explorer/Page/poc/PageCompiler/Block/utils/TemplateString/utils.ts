import { TemplateStringType } from "../../../../schema/primitives/templateString";
import { VariableType } from "../../../../schema/primitives/variable";

/**
 * Regex to match variables enclosed in double curly braces, like {{ variable }}.
 *
 * Explanation:
 * - {{         : Matches the opening double curly braces.
 * - \s*        : Allows optional whitespace after the opening braces.
 * - ([^{}]+?)  : Captures one or more characters that are not { or }.
 * - \s*        : Allows optional whitespace before the closing braces.
 * - }}         : Matches the closing double curly braces literally.
 *
 * The 'g' (global) flag ensures all variable patterns in the string are matched.
 */
export const variablePattern = /{{\s*([^{}]+?)\s*}}/g;

export const replaceVariablesInTemplateString = (
  templateString: TemplateStringType,
  processVariable: (variable: VariableType) => string
) =>
  templateString.replace(variablePattern, (_, variableName) =>
    processVariable(variableName)
  );
