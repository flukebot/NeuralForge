import React from 'react';
import { createRoot } from 'react-dom/client';
import { NodeEditor, Engine } from 'rete';
import { ReactPlugin, Presets, ReactArea2D } from 'rete-react-plugin';
import { AreaPlugin } from 'rete-area-plugin';

class ODiagram extends React.Component {
  constructor(props) {
    super(props);
    this.containerRef = React.createRef();
    this.editor = null;
  }

  async componentDidMount() {
    await this.initializeEditor();
  }

  async initializeEditor() {
    const editor = new NodeEditor('demo@0.1.0', this.containerRef.current);
    const engine = new Engine('demo@0.1.0');

    const render = new ReactPlugin({ createRoot });
    render.addPreset(Presets.classic.setup());

    editor.use(render);
    editor.use(AreaPlugin);

    // Define a simple component
    class AddComponent {
      constructor() {
        this.name = 'Add';
      }

      async builder(node) {
        const numSocket = Presets.classic.Socket.number;
        const inp1 = new Presets.classic.InputControl('num1', numSocket);
        const inp2 = new Presets.classic.InputControl('num2', numSocket);
        const out = new Presets.classic.Output('num', numSocket);

        node.addControl('num1', inp1);
        node.addControl('num2', inp2);
        node.addOutput('num', out);
      }

      worker(node, inputs, outputs) {
        outputs['num'] = (inputs['num1'] || 0) + (inputs['num2'] || 0);
      }
    }

    const component = new AddComponent();
    editor.register(component);
    engine.register(component);

    const node1 = await editor.addNode(component.createNode({ num1: 2, num2: 3 }));
    node1.position = [80, 200];

    const node2 = await editor.addNode(component.createNode({ num1: 5 }));
    node2.position = [400, 200];

    editor.connect(node1.outputs.get('num'), node2.inputs.get('num1'));

    editor.on(
      'process nodecreated noderemoved connectioncreated connectionremoved',
      async () => {
        await engine.abort();
        await engine.process(editor.toJSON());
      }
    );

    editor.trigger('process');
    editor.view.resize();
  }

  render() {
    return (
      <div>
        <h1>ODiagram Visual Programming Interface</h1>
        <div
          ref={this.containerRef}
          style={{ width: '100%', height: '500px', border: '1px solid black' }}
        />
      </div>
    );
  }
}

export default ODiagram;
