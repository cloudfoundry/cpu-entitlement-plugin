require 'sinatra'
require 'concurrent'

$semaphore = Mutex.new
$is_spinning = false

class Spinner < Sinatra::Base
  get '/spin' do
    set_spinning true

    Thread.new do
      while spinning?
	# be nice
	sleep 0.0001
	# spin!
      end
    end

    task = Concurrent::TimerTask.new(execution_interval: params[:spin_time]) do
      set_spinning false
    end

    task.execute
  end

  get '/unspin' do
    set_spinning false
  end

  run! if app_file == $0
end

def spinning?
  $semaphore.synchronize do
     return $is_spinning
  end
end

def set_spinning(spinning)
  $semaphore.synchronize do
     $is_spinning = spinning
  end
end

