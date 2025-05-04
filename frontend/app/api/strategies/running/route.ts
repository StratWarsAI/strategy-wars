import { NextResponse } from 'next/server';

export async function GET() {
  const apiUrl = process.env.BACKEND_API_URL || 'http://localhost:8080/api';
  
  try {
    const response = await fetch(`${apiUrl}/simulations/running`, {
      headers: {
        'Content-Type': 'application/json',
      },
      cache: 'no-store',
    });

    if (!response.ok) {
      throw new Error(`Backend returned ${response.status}: ${response.statusText}`);
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('Error fetching running strategies:', error);
    return NextResponse.json(
      { error: 'Failed to fetch running strategies' },
      { status: 500 }
    );
  }
}