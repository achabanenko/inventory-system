import { Outlet, Link, useLocation } from 'react-router-dom';
import { 
  Package, 
  Warehouse, 
  ShoppingCart, 
  Truck, 
  ClipboardList, 
  Users, 
  Building, 
  MapPin,
  LayoutDashboard,
  LogOut,
  Menu,
  X,
  FolderOpen
} from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';
import { useState } from 'react';
import { cn } from '../lib/utils';

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Items', href: '/items', icon: Package },
  { name: 'Inventory', href: '/inventory', icon: Warehouse },
  { name: 'Purchase Orders', href: '/purchase-orders', icon: ShoppingCart },
  { name: 'Goods Receipts', href: '/receipts', icon: ClipboardList },
  { name: 'Transfers', href: '/transfers', icon: Truck },
  { name: 'Adjustments', href: '/adjustments', icon: ClipboardList },
  { name: 'Stock Counts', href: '/counts', icon: ClipboardList },
  { name: 'Suppliers', href: '/suppliers', icon: Building },
  { name: 'Locations', href: '/locations', icon: MapPin },
  { name: 'Categories', href: '/categories', icon: FolderOpen },
  { name: 'Users', href: '/users', icon: Users, role: 'ADMIN' },
];

export default function Layout() {
  const location = useLocation();
  const { user, tenant, logout } = useAuth();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const filteredNavigation = navigation.filter(
    (item) => !item.role || user?.role === item.role
  );

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Mobile sidebar backdrop */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-gray-600 bg-opacity-75 z-40 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <div
        className={cn(
          'fixed inset-y-0 left-0 z-50 w-64 bg-white shadow-lg transform transition-transform lg:translate-x-0',
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        )}
      >
        <div className="flex h-full flex-col">
          <div className="flex h-16 items-center justify-between px-4 border-b">
            <div>
              <h1 className="text-xl font-bold text-gray-900">Inventory System</h1>
              {tenant && (
                <p className="text-xs text-gray-500">{tenant.name}</p>
              )}
            </div>
            <button
              onClick={() => setSidebarOpen(false)}
              className="lg:hidden"
            >
              <X className="h-6 w-6" />
            </button>
          </div>

          <nav className="flex-1 space-y-1 px-2 py-4">
            {filteredNavigation.map((item) => {
              const isActive = location.pathname === item.href;
              return (
                <Link
                  key={item.name}
                  to={item.href}
                  className={cn(
                    'group flex items-center px-2 py-2 text-sm font-medium rounded-md',
                    isActive
                      ? 'bg-blue-100 text-blue-900'
                      : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                  )}
                >
                  <item.icon
                    className={cn(
                      'mr-3 h-5 w-5',
                      isActive
                        ? 'text-blue-600'
                        : 'text-gray-400 group-hover:text-gray-500'
                    )}
                  />
                  {item.name}
                </Link>
              );
            })}
          </nav>

          <div className="border-t p-4">
            <div className="flex items-center">
              <div className="flex-1">
                <p className="text-sm font-medium text-gray-900">{user?.name}</p>
                <p className="text-xs text-gray-500">{user?.email}</p>
                <p className="text-xs text-gray-500">{user?.role}</p>
              </div>
              <button
                onClick={logout}
                className="ml-3 p-2 text-gray-400 hover:text-gray-500"
              >
                <LogOut className="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="lg:pl-64">
        <div className="sticky top-0 z-10 flex h-16 bg-white shadow lg:hidden">
          <button
            onClick={() => setSidebarOpen(true)}
            className="px-4 text-gray-500 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500"
          >
            <Menu className="h-6 w-6" />
          </button>
          <div className="flex flex-1 items-center justify-center px-4">
            <div className="text-center">
              <h1 className="text-lg font-semibold">Inventory System</h1>
              {tenant && (
                <p className="text-xs text-gray-500">{tenant.name}</p>
              )}
            </div>
          </div>
        </div>

        <main className="p-4 lg:p-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}