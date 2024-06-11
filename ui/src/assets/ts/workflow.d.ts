/**
 * Example method from the example workflow for proof of concept.
 * Needs to be refined.
 * @param params
 */
declare function getFile(params: {
  /**
   * File name
   */
  name: string;
  /**
   * Permission
   */
  permission: number;
  /**
   * What is this for exaclty? What are the possible values?
   */
  scope: "shared" | "other";
}): void;
