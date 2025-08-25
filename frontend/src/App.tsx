import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'react-hot-toast';
import { AuthProvider } from './contexts/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import Layout from './components/Layout';

import Login from './pages/Login';
import Register from './pages/Register';
import TenantLookup from './pages/TenantLookup';
import GoogleOAuthCallback from './pages/GoogleOAuthCallback';
import Dashboard from './pages/Dashboard';
import Items from './pages/Items';
import Inventory from './pages/Inventory';
import PurchaseOrders from './pages/PurchaseOrders';
import Transfers from './pages/Transfers';
import Adjustments from './pages/Adjustments';
import Suppliers from './pages/Suppliers';
import Locations from './pages/Locations';
import Categories from './pages/Categories';
import Users from './pages/Users';
import Counts from './pages/Counts';
import Receipts from './pages/Receipts';
import GoodsReceiptDetails from './components/GoodsReceiptDetails';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5,
      retry: 1,
    },
  },
});

function App() {
  // Debug: Log environment variables on app start
  console.log('=== APP STARTUP DEBUG ===');
  console.log('All environment variables:', import.meta.env);
  console.log('VITE_GOOGLE_CLIENT_ID:', import.meta.env.VITE_GOOGLE_CLIENT_ID);
  console.log('VITE_GOOGLE_REDIRECT_URI:', import.meta.env.VITE_GOOGLE_REDIRECT_URI);
  console.log('VITE_API_URL:', import.meta.env.VITE_API_URL);
  console.log('========================');

  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <Router>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/tenant-lookup" element={<TenantLookup />} />
            <Route path="/google-oauth-callback" element={<GoogleOAuthCallback />} />
            <Route
              path="/"
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route index element={<Navigate to="/dashboard" replace />} />
              <Route path="dashboard" element={<Dashboard />} />
              <Route path="items" element={<Items />} />
              <Route path="inventory" element={<Inventory />} />
              <Route path="purchase-orders" element={<PurchaseOrders />} />
              <Route path="transfers" element={<Transfers />} />
              <Route path="adjustments" element={<Adjustments />} />
              <Route path="counts" element={<Counts />} />
              <Route path="receipts" element={<Receipts />} />
              <Route path="receipts/:id" element={<GoodsReceiptDetails />} />
              <Route path="suppliers" element={<Suppliers />} />
              <Route path="locations" element={<Locations />} />
              <Route path="categories" element={<Categories />} />
              <Route path="users" element={<Users />} />
            </Route>
          </Routes>
        </Router>
        <Toaster />
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App;