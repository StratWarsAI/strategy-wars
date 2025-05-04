import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  const id = params.id;
  
  if (!id || isNaN(Number(id))) {
    return NextResponse.json(
      { error: 'Invalid strategy ID' },
      { status: 400 }
    );
  }
  
  try {
    const API_URL = process.env.BACKEND_API_URL || 'http://localhost:8080/api';
    
    const response = await fetch(`${API_URL}/strategies/${id}`);
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      return NextResponse.json(
        { error: 'Failed to fetch strategy', details: errorData },
        { status: response.status }
      );
    }
    
    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('Error fetching strategy:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}