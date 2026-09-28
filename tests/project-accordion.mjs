// Deliberately limited test consumer, not a production framework converter.
// Reads syntax only. Incoming source is never evaluated or imported.
import ts from 'typescript';

export function projectAccordion(source) {
  const file = ts.createSourceFile('accordion.ts', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
  function reject(node, reason = 'unsupported syntax') {
    const { line, character } = file.getLineAndCharacterOfPosition(node.getStart(file));
    throw new Error(`accordion.ts:${line + 1}:${character + 1}: ${reason}`);
  }
  const ensure = (ok, node, message) => { if (!ok) reject(node, message); };
  if (file.parseDiagnostics.length) reject(file, 'invalid TypeScript');
  const identifier = node => {
    ensure(node && ts.isIdentifier(node), node ?? file, 'expected identifier');
    return node.text;
  };
  function path(node) {
    if (ts.isIdentifier(node)) return node.text;
    if (ts.isPropertyAccessExpression(node)) return path(node.expression) + '.' + identifier(node.name);
    return reject(node, 'expected simple property path');
  }
  function literal(node) {
    if (ts.isAsExpression(node) || ts.isSatisfiesExpression(node)) return literal(node.expression);
    if (node.kind === ts.SyntaxKind.TrueKeyword) return true;
    if (node.kind === ts.SyntaxKind.FalseKeyword) return false;
    if (ts.isStringLiteral(node)) return node.text;
    if (ts.isNumericLiteral(node)) return Number(node.text);
    if (ts.isObjectLiteralExpression(node)) {
      const result = {};
      for (const p of node.properties) {
        ensure(ts.isPropertyAssignment(p), p, 'expected literal property');
        const key = identifier(p.name);
        ensure(!Object.hasOwn(result, key), p, 'duplicate property');
        Object.defineProperty(result, key, { value: literal(p.initializer), enumerable: true });
      }
      return result;
    }
    return reject(node, 'expected literal');
  }
  const exports = new Map();
  for (const statement of file.statements) {
    ensure(statement.modifiers?.some(m => m.kind === ts.SyntaxKind.ExportKeyword), statement, 'expected exported declaration');
    let name, value;
    if (ts.isVariableStatement(statement)) {
      ensure(statement.declarationList.flags & ts.NodeFlags.Const, statement, 'expected const');
      ensure(statement.declarationList.declarations.length === 1, statement);
      const declaration = statement.declarationList.declarations[0];
      name = identifier(declaration.name);
      value = literal(declaration.initializer);
    } else if (ts.isInterfaceDeclaration(statement) || ts.isFunctionDeclaration(statement)) {
      name = identifier(statement.name); value = statement;
    } else reject(statement);
    ensure(!exports.has(name), statement, 'duplicate export');
    exports.set(name, value);
  }
  const expectedExports = ['contractVersion','AccordionProps','defaults','AccordionSlots','AccordionEvents','nativeEvents','accordion'];
  ensure(exports.size === expectedExports.length && expectedExports.every(k => exports.has(k)), file, 'unknown or missing export');
  ensure(exports.get('contractVersion') === 1, file, 'unsupported contract version');
  const props = {};
  for (const member of exports.get('AccordionProps').members) {
    ensure(ts.isPropertySignature(member) && member.type, member, 'expected prop signature');
    props[identifier(member.name)] = { type: member.type.getText(file), optional: !!member.questionToken };
  }
  const slotsInterface = exports.get('AccordionSlots');
  ensure(slotsInterface.typeParameters?.length === 1 && !slotsInterface.typeParameters[0].constraint, slotsInterface, 'content must be replaceable');
  const slots = {};
  for (const member of slotsInterface.members) {
    ensure(ts.isPropertySignature(member) && ts.isFunctionTypeNode(member.type), member, 'expected slot signature');
    const scope = {};
    ensure(member.type.parameters.length <= 1, member);
    for (const parameter of member.type.parameters) {
      ensure(identifier(parameter.name) === 'scope' && ts.isTypeLiteralNode(parameter.type), parameter);
      for (const field of parameter.type.members) {
        ensure(ts.isPropertySignature(field) && field.type, field);
        scope[identifier(field.name)] = field.type.getText(file);
      }
    }
    slots[identifier(member.name)] = { optional: !!member.questionToken, scope };
  }
  const nodes = new Map();
  const bound = new Set();
  let returned = false;
  const factory = exports.get('accordion');
  ensure(factory.body && factory.parameters.length === 2 && !factory.asteriskToken && !factory.typeParameters, factory);
  ensure(factory.modifiers?.every(m => m.kind === ts.SyntaxKind.ExportKeyword), factory, 'unsupported factory modifier');
  for (const parameter of factory.parameters) ensure(!parameter.initializer && !parameter.dotDotDotToken && !parameter.questionToken && !parameter.modifiers, parameter, 'unsupported factory parameter');
  ensure(factory.parameters.map(p => identifier(p.name)).join(',') === 'props,slots', factory);
  function prop(node) {
    const name = identifier(node);
    ensure(bound.has(name), node, 'unknown prop binding');
    return name;
  }
  function nodeByName(name, location) {
    ensure(nodes.has(name), location, 'unknown native node');
    return nodes.get(name);
  }
  function expression(statement, guard) {
    const expr = statement.expression;
    if (ts.isBinaryExpression(expr) && expr.operatorToken.kind === ts.SyntaxKind.EqualsToken) {
      ensure(ts.isPropertyAccessExpression(expr.left), expr.left);
      const target = nodeByName(identifier(expr.left.expression), expr.left);
      const name = identifier(expr.left.name), input = prop(expr.right);
      ensure(guard === (props[input].optional && !Object.hasOwn(exports.get('defaults'), input) ? input : undefined), expr, 'incorrect optional guard');
      target.bindings.push({ kind: 'property', name, prop: input, optional: !!guard });
      return;
    }
    ensure(ts.isCallExpression(expr) && ts.isPropertyAccessExpression(expr.expression), expr, 'unsupported call');
    const target = nodeByName(identifier(expr.expression.expression), expr);
    const method = identifier(expr.expression.name);
    if (method === 'setAttribute') {
      ensure(expr.arguments.length === 2 && ts.isStringLiteral(expr.arguments[0]), expr);
      const input = prop(expr.arguments[1]);
      ensure(guard === (props[input].optional && !Object.hasOwn(exports.get('defaults'), input) ? input : undefined), expr, 'incorrect optional guard');
      target.bindings.push({ kind: 'attribute', name: expr.arguments[0].text, prop: input, optional: !!guard });
    } else if (method === 'append') {
      ensure(expr.arguments.length === 1, expr);
      const arg = expr.arguments[0];
      if (ts.isIdentifier(arg)) {
        ensure(!guard, expr);
        target.children.push(nodeByName(identifier(arg), arg));
      } else {
        ensure(ts.isCallExpression(arg) && ts.isPropertyAccessExpression(arg.expression), arg);
        ensure(identifier(arg.expression.expression) === 'slots', arg);
        const name = identifier(arg.expression.name), scope = {};
        ensure(Object.hasOwn(slots, name), arg, 'unknown slot');
        ensure(guard === (slots[name].optional ? 'slots.' + name : undefined), arg, 'incorrect slot guard');
        ensure(arg.arguments.length <= 1, arg);
        if (arg.arguments.length) {
          const object = arg.arguments[0];
          ensure(ts.isObjectLiteralExpression(object), object);
          for (const field of object.properties) {
            ensure(ts.isShorthandPropertyAssignment(field) && !field.objectAssignmentInitializer, field, 'unsupported scope expression');
            const key = identifier(field.name); scope[key] = prop(field.name);
          }
        }
        ensure(Object.keys(scope).sort().join(',') === Object.keys(slots[name].scope).sort().join(','), arg, 'incorrect slot scope');
        target.children.push({ slot: name, scope, optional: !!guard });
      }
    } else reject(expr, 'unsupported method');
  }
  for (const statement of factory.body.statements) {
    ensure(!returned, statement, 'statement after return');
    if (ts.isVariableStatement(statement)) {
      ensure(statement.declarationList.flags & ts.NodeFlags.Const, statement);
      ensure(statement.declarationList.declarations.length === 1, statement);
      const d = statement.declarationList.declarations[0];
      if (ts.isObjectBindingPattern(d.name)) {
        ensure(identifier(d.initializer) === 'props', d);
        for (const field of d.name.elements) {
          const name = identifier(field.name);
          ensure(!field.propertyName && !field.dotDotDotToken && Object.hasOwn(props, name) && !bound.has(name), field);
          if (field.initializer) ensure(path(field.initializer) === 'defaults.' + name && Object.hasOwn(exports.get('defaults'), name), field);
          else ensure(!Object.hasOwn(exports.get('defaults'), name), field, 'missing default initialization');
          bound.add(name);
        }
      } else {
        const name = identifier(d.name), call = d.initializer;
        ensure(ts.isCallExpression(call) && path(call.expression) === 'document.createElement' && call.arguments.length === 1 && ts.isStringLiteral(call.arguments[0]), d, 'expected literal native element creation');
        ensure(!nodes.has(name), d, 'duplicate node');
        nodes.set(name, { tag: call.arguments[0].text, bindings: [], children: [] });
      }
    } else if (ts.isExpressionStatement(statement)) expression(statement);
    else if (ts.isIfStatement(statement)) {
      const condition = statement.expression;
      ensure(!statement.elseStatement && ts.isExpressionStatement(statement.thenStatement), statement);
      ensure(ts.isBinaryExpression(condition) && condition.operatorToken.kind === ts.SyntaxKind.ExclamationEqualsEqualsToken && identifier(condition.right) === 'undefined', condition);
      expression(statement.thenStatement, path(condition.left));
    } else if (ts.isReturnStatement(statement)) {
      ensure(identifier(statement.expression) === 'root' && nodes.has('root'), statement);
      returned = true;
    } else reject(statement);
  }
  ensure(returned && bound.size === Object.keys(props).length, factory, 'incomplete factory');
  return { props, defaults: exports.get('defaults'), slots, events: exports.get('nativeEvents'), tree: nodes.get('root') };
}
