import {EditorView, basicSetup} from "codemirror"
import { EditorState } from '@codemirror/state';
import {parser} from "./strawk.parser.js"
import {foldNodeProp, foldInside, indentNodeProp} from "@codemirror/language"
import {styleTags, tags as t} from "@lezer/highlight"
import {LRLanguage} from "@codemirror/language"
import {completeFromList} from "@codemirror/autocomplete"
import {LanguageSupport} from "@codemirror/language"

export const parserWithMetadata = parser.configure({
  props: [
    styleTags({
      "BEGIN END do while for in continue print next if else break length sub gsub split toupper tolower substr" : t.keyword,
      identifier: t.tagName,
      stateidentifier: t.variableName,
      String: t.string,
      Regex: t.regexp,
      Boolean: t.bool,
      String: t.string,
      LineComment: t.lineComment
    }),
    indentNodeProp.add({
      Application: context => context.column(context.node.from) + context.unit
    }),
    foldNodeProp.add({
      Application: foldInside
    })
  ]
});



export const exampleLanguage = LRLanguage.define({
  name: "strawk",
  parser: parserWithMetadata,
  languageData: {
        commentTokens: {line: "#"}
  }
});

export function strawk() {
    return new LanguageSupport(exampleLanguage,  []);
}

function createEditorStateForStrawk(initialContents, options = {}) {
    let extensions = [
      basicSetup,
      strawk()
    ];

    return EditorState.create({
        doc: initialContents,
        extensions
    });
}

function createEditorState(initialContents, options = {}) {
    let extensions = [
      basicSetup
    ];

    return EditorState.create({
        doc: initialContents,
        extensions
    });
}

function createEditorView(state, parent) {
    return new EditorView({ state, parent });
}

export { createEditorStateForStrawk, createEditorState, createEditorView};


