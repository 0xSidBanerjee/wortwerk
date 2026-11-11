/* eslint-disable */
/** Auto-generated — run `npm run generate-types`. Do not edit. */

export interface WortwerkCompiledStory {
  schemaVersion: 1;
  id: string;
  title: string;
  lang: string;
  glossLang: string;
  level: string;
  wordCount: number;
  sentences: Sentence[];
  lexicon: {
    [k: string]: LexiconEntry;
  };
}
export interface Sentence {
  id: string;
  p: number;
  /**
   * @minItems 1
   */
  tokens: [Token, ...Token[]];
  units: Unit[];
  translation: string;
}
export interface Token {
  id: number;
  text: string;
  space: boolean;
}
export interface Unit {
  /**
   * @minItems 1
   */
  tokens: [number, ...number[]];
  lex: string;
  form?: string;
  gloss?: string;
}
export interface LexiconEntry {
  lemma: string;
  kind: "word" | "phrase" | "separable" | "compound" | "name";
  pos: string;
  meaning: string;
  gender?: "m" | "f" | "n";
  article?: string;
  plural?: string | null;
  separable?: Separable;
  /**
   * @minItems 1
   */
  parts?: [string, ...string[]];
  audio?: string;
}
export interface Separable {
  prefix: string;
  stem: string;
}
