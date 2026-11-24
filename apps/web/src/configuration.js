export const targetKeys = ['api', 'web', 'ios', 'android', 'wechat', 'alipay', 'admin'];

export const stackFamilies = [
  {key: 'apiStack', flag: 'api-stack', label: '后端技术栈', targets: ['api'], options: [
    ['go', 'Go', 'Go 标准库 API 与 OpenAPI'],
    ['nestjs', 'NestJS', 'NestJS / TypeScript API 与 OpenAPI']
  ]},
  {key: 'webStack', flag: 'web-stack', label: 'Web 技术栈', targets: ['web'], options: [
    ['react', 'React', 'React / TypeScript 与 Vite 构建'],
    ['vue', 'Vue', 'Vue 3 / TypeScript 与 Vite 构建']
  ]},
  {key: 'mobileStack', flag: 'mobile-stack', label: '移动端技术栈', targets: ['ios', 'android'], options: [
    ['react-native', 'React Native / Expo', 'iOS 与 Android 共用 Expo 项目；构建输出 JS 资源，原生签名另行配置。'],
    ['flutter', 'Flutter', 'iOS 与 Android 共用 Flutter 项目；需 Flutter SDK、对应平台工具链及发布签名。']
  ]},
  {key: 'miniappStack', flag: 'miniapp-stack', label: '小程序技术栈', targets: ['wechat', 'alipay'], options: [
    ['taro', 'Taro', 'Taro / React；平台 App ID、开发者工具及审核另行配置。'],
    ['uniapp', 'uni-app', 'uni-app / Vue 3；平台 App ID、开发者工具及审核另行配置。']
  ]},
  {key: 'adminStack', flag: 'admin-stack', label: 'Admin 模板', targets: ['admin'], options: [
    ['vben', 'Vben Naive UI', 'Vben / Vue 3 / Naive UI；独立 pnpm 工作区，需 Node.js 22.18+，使用 make run-admin 启动。登录与权限为模板演示，需接入自己的后端。'],
    ['soybean', 'Soybean-admin', 'Soybean / Vue 3 / Naive UI；独立 pnpm 工作区，需 Node.js 22.18+，使用 make run-admin 启动。登录与权限为模板演示，需接入自己的后端。']
  ]}
];

export const defaultStacks = Object.fromEntries(stackFamilies.map(family => [family.key, family.options[0][0]]));

export function projectConfiguration(name, targets, stacks, visual, openSpec) {
  const selectedTargets = targetKeys.filter(key => targets[key]);
  const enabledFamilies = stackFamilies.filter(family => family.targets.some(key => targets[key]));
  const selectedStacks = Object.fromEntries(enabledFamilies.map(family => [family.key, stacks[family.key]]));
  return {
    config: {name, ...targets, ...selectedStacks, visual, openspec: openSpec},
    command: `aistd create --name ${name} --targets ${selectedTargets.join(',')}${enabledFamilies.map(family => ` --${family.flag} ${stacks[family.key]}`).join('')}${visual ? '' : ' --visual=false'}${openSpec ? ' --openspec' : ''}`
  };
}
