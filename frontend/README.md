# HSA Receipt Manager - Frontend

A modern Vue 3 web application for managing Health Savings Account (HSA) receipts with receipt upload, viewing, editing, and deduction calculation capabilities.

## Features

- **📤 Receipt Upload**: Drag-and-drop or click to upload receipt images (JPG, PNG, HEIC, PDF)
- **📋 Receipt Management**: View, edit, and delete receipts with comprehensive details
- **🧮 Deduction Calculator**: Find optimal receipt combinations to match target HSA reimbursement amounts
- **📊 Real-time Summary**: Track available vs. used HSA amounts
- **🖼️ Image Preview**: View receipt images directly in the browser with full-size viewing option
- **✅ HSA Status Tracking**: Mark receipts as Yes/No/Partially HSA-qualified
- **📝 Usage Tracking**: Record when and why receipts were used for reimbursement

## Tech Stack

- **Framework**: Vue 3 (Composition API)
- **UI Library**: Vuetify 3
- **Routing**: Vue Router 4
- **State Management**: Pinia
- **HTTP Client**: Axios
- **Build Tool**: Vite
- **Package Manager**: pnpm (or npm)
- **Containerization**: Docker with Nginx

## Project Structure

```
frontend/
├── src/
│   ├── router/
│   │   └── index.js           # Route definitions
│   ├── services/
│   │   └── api.js             # API client and HTTP requests
│   ├── stores/
│   │   └── receipts.js        # Pinia store for receipt state
│   ├── views/
│   │   ├── HSADeduction.vue   # Deduction calculator page
│   │   ├── ReceiptList.vue    # Receipt management page
│   │   └── UploadReceipt.vue  # Receipt upload page
│   ├── App.vue                # Root component
│   └── main.js                # Application entry point
├── Dockerfile                 # Production Docker build
├── Dockerfile.dev             # Development Docker build
├── Dockerfile.prod            # Alternate production build
├── nginx.conf                 # Nginx configuration for SPA
├── index.html                 # HTML template
├── package.json               # Dependencies and scripts
├── pnpm-lock.yaml            # Lock file for pnpm
└── README.md                  # This file
```

## Prerequisites

- Node.js 18+ (for local development)
- pnpm, npm, or yarn
- Docker (for containerized deployment)

## Environment Variables

Create a `.env` file in the root directory:

```bash
# API Backend URL
# For local development with backend on localhost:8080
VITE_API_URL=http://localhost:8080/api

# For production, leave empty to auto-detect from current host
# VITE_API_URL=

# Optional: Custom API port for dynamic URL construction
# VITE_API_PORT=30081
```

### Environment Variable Details

- **`VITE_API_URL`**: Full URL to the API backend
  - If set, this URL will be used for all API requests
  - If not set, the app will construct the URL dynamically based on the current hostname
  - Format: `http://hostname:port/api` or `https://hostname/api`

- **`VITE_API_PORT`**: (Optional) Port number for API when using dynamic URL construction
  - Only used when `VITE_API_URL` is not set
  - Default: `30081`

## Installation

### Local Development

```bash
# Install dependencies
pnpm install
# or
npm install

# Start development server
pnpm dev
# or
npm run dev
```

The application will be available at `http://localhost:5173`

### Production Build

```bash
# Build for production
pnpm build
# or
npm run build

# Preview production build locally
pnpm preview
# or
npm run preview
```

## Docker Deployment

### Development Container

```bash
# Build development image
docker build -f Dockerfile.dev -t hsa-frontend-dev .

# Run development container
docker run -p 5173:5173 \
  -v $(pwd):/app \
  -v /app/node_modules \
  hsa-frontend-dev
```

### Production Container

```bash
# Build production image
docker build -t hsa-frontend .

# Run production container
docker run -p 80:80 hsa-frontend
```

### Production with Custom API URL

```bash
# Build with custom API URL
docker build \
  --build-arg VITE_API_URL=http://api.example.com/api \
  -t hsa-frontend \
  -f Dockerfile.prod \
  .

# Run
docker run -p 80:80 hsa-frontend
```

## Available Scripts

| Script | Description |
|--------|-------------|
| `pnpm dev` | Start development server with hot reload |
| `pnpm build` | Build for production |
| `pnpm preview` | Preview production build locally |

## Application Routes

| Path | Component | Description |
|------|-----------|-------------|
| `/` | Redirect | Redirects to `/upload` |
| `/upload` | UploadReceipt | Upload new receipts |
| `/receipts` | ReceiptList | View and manage all receipts |
| `/deduct` | HSADeduction | Calculate optimal receipt combinations |

## Key Features Explained

### Receipt Upload
- Supports drag-and-drop and click-to-upload
- Accepts JPG, PNG, HEIC, and PDF files
- Shows image preview before upload
- Automatic duplicate detection
- Real-time OCR processing through backend API

### Receipt Management
- View all receipts in a sortable data table
- Edit receipt details (vendor, amount, date, HSA status)
- View receipt images directly in browser
- Mark receipts as used/unused
- Add usage reason for tracking
- Delete receipts with confirmation
- Reference guide for common HSA-qualified expenses

### HSA Status Options
- **Yes**: Fully HSA-qualified
- **No**: Not HSA-qualified
- **Partially**: Partially qualified (with proportional tax calculation guide)

### Deduction Calculator
- Input target reimbursement amount
- Algorithm finds optimal receipt combination
- Shows total selected amount
- Approve and mark receipts as used in one action
- Tracks available balance in real-time

## API Integration

The frontend communicates with the Go backend API through the following endpoints:

- `POST /api/receipts/upload` - Upload new receipt
- `GET /api/receipts` - Get all receipts
- `GET /api/receipts/{id}` - Get specific receipt
- `PUT /api/receipts/{id}` - Update receipt
- `DELETE /api/receipts/{id}` - Delete receipt
- `POST /api/receipts/deduct` - Calculate optimal receipt combination
- `GET /receipts/file/{id}` - Serve receipt image file

## Nginx Configuration

The included `nginx.conf` provides:

- SPA routing support (all routes fall back to `index.html`)
- API proxy to backend service
- Static asset caching with 1-year expiry
- Gzip compression for text assets
- Receipt file serving through proxy

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
- Mobile browsers (iOS Safari, Chrome Mobile)

## Responsive Design

The application is fully responsive and works on:
- Desktop (1920px+)
- Laptop (1280px - 1919px)
- Tablet (768px - 1279px)
- Mobile (320px - 767px)

## Development Tips

### Hot Module Replacement (HMR)
Vite provides instant HMR during development. Changes to `.vue` files are reflected immediately.

### Vue Devtools
Install [Vue Devtools](https://devtools.vuejs.org/) browser extension for debugging:
- Inspect component hierarchy
- View Pinia store state
- Track router navigation
- Monitor performance

### API URL Configuration
When developing locally:
1. Set `VITE_API_URL=http://localhost:8080/api` in `.env`
2. Ensure backend API is running on port 8080
3. CORS is handled by the backend

### Debugging API Calls
The `api.js` service includes console logging for all requests. Check browser console for:
- API endpoint URLs
- Request/response data
- Error details

## Common Issues

### API Connection Errors

**Problem**: Frontend cannot connect to backend API

**Solutions**:
1. Verify `VITE_API_URL` is set correctly
2. Check backend API is running
3. Ensure CORS is enabled on backend
4. Check browser console for actual request URL

### Images Not Loading

**Problem**: Receipt images don't display

**Solutions**:
1. Verify image path in database matches file system
2. Check nginx proxy configuration for `/receipts/` path
3. Ensure backend file serving endpoint is working
4. Check browser network tab for 404 errors

### Build Failures

**Problem**: `pnpm build` fails

**Solutions**:
1. Clear node_modules and reinstall: `rm -rf node_modules && pnpm install`
2. Clear Vite cache: `rm -rf node_modules/.vite`
3. Check for TypeScript/ESLint errors
4. Ensure all dependencies are compatible

## Performance Optimization

### Production Build
- Code splitting for optimal loading
- Tree shaking removes unused code
- Assets minified and compressed
- Lazy loading for routes

### Caching Strategy
- Static assets cached for 1 year
- API responses not cached (real-time data)
- Service worker can be added for offline support

## Security Considerations

- No sensitive data stored in frontend code
- API URL configuration via environment variables
- HTTPS recommended for production
- CORS handled by backend
- No authentication tokens exposed in browser storage (add auth as needed)

## Future Enhancements

- [ ] Add user authentication and multi-user support
- [ ] Implement receipt categories and tags
- [ ] Add export functionality (CSV, PDF reports)
- [ ] Implement receipt search and filtering
- [ ] Add receipt analytics and spending trends
- [ ] Support for batch receipt upload
- [ ] Dark mode theme
- [ ] Offline support with service worker
- [ ] Receipt OCR accuracy improvement feedback

## Support

For issues or questions:
- Check the [API documentation](../api-go/README.md)
- Review browser console for errors
- Verify environment configuration
- Check Docker logs: `docker logs <container-id>`