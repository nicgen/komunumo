export default {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [
      2,
      'always',
      ['feat', 'fix', 'docs', 'chore', 'refactor', 'test', 'perf', 'build', 'ci', 'style'],
    ],
    'scope-enum': [
      2,
      'always',
      ['auth', 'posts', 'chat', 'notif', 'profile', 'follows', 'groups', 'events', 'search', 'audit', 'rgpd', 'db', 'api', 'web', 'ws', 'ops', 'adr', 'specs', 'docs', 'ci', 'deps', 'scaffold', 'release', 'learnings'],
    ],
    'subject-case': [2, 'never', ['upper-case', 'pascal-case']],
    'header-max-length': [2, 'always', 100],
    'body-max-line-length': [1, 'always', 120],
  },
};
