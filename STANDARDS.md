# DevSmith Coding Standards

Based on patterns from DevSmith Logs project.

## File Organization

### Backend (Python/FastAPI)
```
backend/
├── main.py              # FastAPI app
├── models/              # Database models
├── routes/              # API endpoints
├── services/            # Business logic
├── tests/               # Test files
└── requirements.txt     # Dependencies
```

### Frontend (React)
```
frontend/
├── src/
│   ├── components/      # Reusable components
│   ├── context/         # Context providers
│   ├── hooks/           # Custom hooks
│   ├── styles/          # CSS files
│   ├── utils/           # Utility functions
│   └── __tests__/       # Test files
├── public/
└── package.json
```

## Naming Conventions

### Files
- React Components: `PascalCase.jsx` (e.g., `UserProfile.jsx`)
- Utilities: `camelCase.js` (e.g., `apiClient.js`)
- Styles: `kebab-case.css` (e.g., `user-profile.css`)
- Tests: `ComponentName.test.jsx`

### Code
- Variables: `camelCase`
- Constants: `UPPER_SNAKE_CASE`
- Functions: `camelCase`
- Classes/Components: `PascalCase`

## React Component Structure

```javascript
import React, { useState, useEffect } from 'react';
import { useContext } from 'react';

// Component with proper prop destructuring and defaults
export default function MyComponent({
  prop1,
  prop2 = 'default',
  onAction
}) {
  // State
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Context
  const { theme } = useTheme();

  // Effects
  useEffect(() => {
    // Effect logic
  }, []);

  // Handlers
  const handleClick = () => {
    // Handler logic
  };

  // Render guards
  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div className={`component ${theme}`}>
      {/* JSX */}
    </div>
  );
}
```

## API Call Pattern

```javascript
async function fetchData(params) {
  try {
    const response = await fetch(url, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' }
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const data = await response.json();
    return data.key || fallbackValue;
  } catch (err) {
    console.error('Error:', err);
    return fallbackValue;
  }
}
```

## Error Handling

Always provide:
1. User-friendly error messages
2. Fallback values
3. Loading states
4. Console logging for debugging

## Testing Requirements

- Unit tests for utilities
- Component tests for React
- API endpoint tests for backend
- Integration tests for critical paths
- Minimum 70% coverage

## Git Workflow

1. Create feature branch from `development`
2. Make changes
3. Update appropriate changelog
4. Test locally
5. Commit with descriptive message
6. Push and create PR to `development`
7. After approval, merge to `development`
8. Release: merge `development` to `main`

---
