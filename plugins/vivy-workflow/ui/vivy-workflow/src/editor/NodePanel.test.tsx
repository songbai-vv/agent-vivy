// @vitest-environment happy-dom
import { afterEach, describe, expect, it, vi } from 'vitest';
import { act } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { NodePanel } from './NodePanel';
import type { Artifact, Node } from '../studio/schema';

(globalThis as { IS_REACT_ACT_ENVIRONMENT?: boolean }).IS_REACT_ACT_ENVIRONMENT = true;

function artifactWith(node: Node): Artifact {
  return {
    definition: {
      schema_version: 'inofy.workflow/v1',
      graph: {
        nodes: [
          { id: 'research', kind: 'call', type: 'vivy.child-task@1', config: { task: 'Research' } },
          node,
        ],
        edges: [{ from: 'research', to: node.id }],
        exits: [node.id],
      },
    },
  } as Artifact;
}

let root: Root | null = null;
let container: HTMLElement | null = null;

afterEach(() => {
  act(() => root?.unmount());
  container?.remove();
  root = null;
  container = null;
});

function commitInputs(node: Node, text: string): Node {
  const artifact = artifactWith(node);
  const onChange = vi.fn();
  container = document.createElement('div');
  document.body.appendChild(container);
  root = createRoot(container);
  act(() => {
    root!.render(
      <NodePanel artifact={artifact} node={node} descriptor={undefined} t={(key: string) => key} onChange={onChange} onDelete={vi.fn()} />,
    );
  });
  const textarea = container.querySelector<HTMLTextAreaElement>('#wf-node-inputs');
  expect(textarea).not.toBeNull();
  // The panel commits on blur and reads the textarea's value at that point.
  textarea!.value = text;
  act(() => {
    textarea!.dispatchEvent(new FocusEvent('focusout', { bubbles: true }));
  });
  const next = onChange.mock.calls.at(-1)![0] as Artifact;
  return next.definition.graph.nodes.find((n) => n.id === node.id)!;
}

const binding = { research: { source: 'research', pointer: '/result' } };

describe('NodePanel inputs', () => {
  it('writes authored bindings to node.inputs, not config.inputs', () => {
    const node = commitInputs(
      { id: 'summary', kind: 'call', type: 'vivy.child-task@1', config: { task: 'Summarize' } },
      JSON.stringify(binding),
    );
    expect(node.inputs).toEqual(binding);
    expect(node.config).toEqual({ task: 'Summarize' });
  });

  it('moves a draft that still carries config.inputs onto node.inputs', () => {
    const node = commitInputs(
      { id: 'summary', kind: 'call', type: 'vivy.child-task@1', config: { task: 'Summarize', inputs: binding } },
      JSON.stringify(binding),
    );
    expect(node.inputs).toEqual(binding);
    expect(node.config).toEqual({ task: 'Summarize' });
  });

  it('drops node.inputs when the bindings are cleared', () => {
    const node = commitInputs(
      { id: 'summary', kind: 'call', type: 'vivy.child-task@1', config: { task: 'Summarize' }, inputs: binding },
      '{}',
    );
    expect(node).not.toHaveProperty('inputs');
  });
});
