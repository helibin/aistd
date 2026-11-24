import test from 'node:test';
import assert from 'node:assert/strict';
import {defaultStacks, projectConfiguration, targetKeys} from './configuration.js';

const targets = selected => Object.fromEntries(targetKeys.map(key => [key, selected.includes(key)]));

test('default configuration explicitly selects Go and React only', () => {
  const result = projectConfiguration('demo', targets(['api', 'web']), defaultStacks, true, false);
  assert.equal(result.command, 'aistd create --name demo --targets api,web --api-stack go --web-stack react');
  assert.equal(result.config.apiStack, 'go');
  assert.equal(result.config.webStack, 'react');
  assert.equal('mobileStack' in result.config, false);
  assert.equal('miniappStack' in result.config, false);
  assert.equal(result.config.admin, false);
  assert.equal('adminStack' in result.config, false);
});

test('all alternative stacks agree between command and JSON', () => {
  const stacks = {apiStack: 'nestjs', webStack: 'vue', mobileStack: 'flutter', miniappStack: 'uniapp', adminStack: 'soybean'};
  const result = projectConfiguration('demo', targets(targetKeys), stacks, false, true);
  assert.equal(result.command, 'aistd create --name demo --targets api,web,ios,android,wechat,alipay,admin --api-stack nestjs --web-stack vue --mobile-stack flutter --miniapp-stack uniapp --admin-stack soybean --visual=false --openspec');
  assert.deepEqual(result.config, {name: 'demo', ...targets(targetKeys), ...stacks, visual: false, openspec: true});
});

test('Admin templates generate independent commands and JSON', () => {
  for (const adminStack of ['vben', 'soybean']) {
    const result = projectConfiguration('hello-aistd', targets(['admin']), {...defaultStacks, adminStack}, true, false);
    assert.equal(result.command, `aistd create --name hello-aistd --targets admin --admin-stack ${adminStack}`);
    assert.deepEqual(result.config, {name: 'hello-aistd', ...targets(['admin']), adminStack, visual: true, openspec: false});
  }
});

test('disabled Admin omits its stack and restores the previous selection when enabled', () => {
  const stacks = {...defaultStacks, adminStack: 'soybean'};
  const disabled = projectConfiguration('demo', targets(['web']), stacks, true, false);
  assert.equal('adminStack' in disabled.config, false);
  assert.equal(disabled.command.includes('--admin-stack'), false);
  const restored = projectConfiguration('demo', targets(['admin']), stacks, true, false);
  assert.equal(restored.config.adminStack, 'soybean');
  assert.equal(defaultStacks.adminStack, 'vben');
});

test('shared family remains selected until its last target is removed', () => {
  const stacks = {...defaultStacks, mobileStack: 'flutter', miniappStack: 'uniapp'};
  const single = projectConfiguration('demo', targets(['android', 'alipay']), stacks, true, false);
  assert.equal(single.config.mobileStack, 'flutter');
  assert.equal(single.config.miniappStack, 'uniapp');
  assert.equal(single.command.includes('--api-stack'), false);
  assert.equal(single.command.includes('--web-stack'), false);
  const disabled = projectConfiguration('demo', targets([]), stacks, true, false);
  assert.equal('mobileStack' in disabled.config, false);
  assert.equal('miniappStack' in disabled.config, false);
  assert.equal(disabled.command.includes('-stack'), false);
  const restored = projectConfiguration('demo', targets(['ios', 'wechat']), stacks, true, false);
  assert.equal(restored.config.mobileStack, 'flutter');
  assert.equal(restored.config.miniappStack, 'uniapp');
});
