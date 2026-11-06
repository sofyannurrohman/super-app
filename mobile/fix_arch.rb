require 'xcodeproj'
project_path = 'ios/Runner.xcodeproj'
project = Xcodeproj::Project.open(project_path)
project.build_configurations.each do |config|
  config.build_settings['EXCLUDED_ARCHS[sdk=iphonesimulator*]'] = 'arm64'
end
project.save
