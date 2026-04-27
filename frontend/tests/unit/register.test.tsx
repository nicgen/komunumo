import { render } from '@testing-library/react';
import { axe, toHaveNoViolations } from 'jest-axe';
import { expect, it, describe } from 'vitest';
import RegisterPage from '../../app/(auth)/register/page';

expect.extend(toHaveNoViolations);

describe('RegisterPage a11y', () => {
  it('has no axe-core violations', async () => {
    const { container } = render(<RegisterPage />);
    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });
});
