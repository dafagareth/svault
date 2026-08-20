<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\Task;
use Illuminate\Http\Request;

class TaskController extends Controller
{
    public function index(Request $request)
    {
        $query = Task::forUser($request->user()->id)
            ->latest();

        if ($request->filled('status')) {
            $query->where('status', $request->status);
        }

        if ($request->filled('priority')) {
            $query->where('priority', $request->priority);
        }

        if ($request->filled('search')) {
            $query->where(function ($q) use ($request) {
                $q->where('title', 'like', "%{$request->search}%")
                  ->orWhere('description', 'like', "%{$request->search}%");
            });
        }

        return response()->json($query->paginate(20));
    }

    public function store(Request $request)
    {
        $data = $request->validate([
            'title'       => 'required|string|max:255',
            'description' => 'nullable|string',
            'status'      => 'in:todo,in_progress,done',
            'priority'    => 'in:low,medium,high',
            'due_date'    => 'nullable|date',
        ]);

        $task = $request->user()->tasks()->create($data);

        return response()->json($task, 201);
    }

    public function show(Request $request, Task $task)
    {
        $this->authorize($request->user(), $task);

        return response()->json($task);
    }

    public function update(Request $request, Task $task)
    {
        $this->authorize($request->user(), $task);

        $data = $request->validate([
            'title'       => 'sometimes|string|max:255',
            'description' => 'nullable|string',
            'status'      => 'sometimes|in:todo,in_progress,done',
            'priority'    => 'sometimes|in:low,medium,high',
            'due_date'    => 'nullable|date',
        ]);

        $task->update($data);

        return response()->json($task);
    }

    public function destroy(Request $request, Task $task)
    {
        $this->authorize($request->user(), $task);
        $task->delete();

        return response()->noContent();
    }

    private function authorize(object $user, Task $task): void
    {
        abort_if($task->user_id !== $user->id, 403, 'Forbidden');
    }
}
