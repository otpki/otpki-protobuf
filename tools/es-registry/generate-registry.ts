import { writeFile } from 'fs/promises';
import { glob } from 'glob';

const pbFiles = (await glob('es/gen/public/**/*_pb.js')).sort();
const imports = pbFiles.map((file, i) =>
  `import * as mod${i} from './${file.replace('es/gen/', '')}';`
).join('\n');

const fileDescs = pbFiles.map((_, i) =>
  `...Object.entries(mod${i}).filter(([k]) => k.startsWith('file_')).map(([, v]) => v)`
).join(',\n  ');

const tsCode = `import { createRegistry, type Registry } from '@bufbuild/protobuf';

${imports}

const fileDescs = [
  ${fileDescs}
];

export const registry: Registry = createRegistry(...fileDescs);
`;

const jsCode = `import { createRegistry } from '@bufbuild/protobuf';

${imports}

const fileDescs = [
  ${fileDescs}
];

export const registry = createRegistry(...fileDescs);
`;

const dtsCode = `import type { Registry } from '@bufbuild/protobuf';

export declare const registry: Registry;
`;

await Promise.all([
  writeFile('es/gen/registry.ts', tsCode),
  writeFile('es/gen/registry.js', jsCode),
  writeFile('es/gen/registry.d.ts', dtsCode),
]);
