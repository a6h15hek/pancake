#!/bin/bash

# Define the functions
status() { echo "📊 Status function called"; }

project_list() {
    echo "📚 Project list:"
    printf "| %-10s | %-10s | %-20s | %-10s | %-30s |\n" "Name" "Branch" "Last Committer" "Version" "Last Updated"
    echo "|------------|------------|----------------------|------------|--------------------------------|"
    # Parse pancake.yml and list each project
    project_location=$(yq e '.project_location' pancake.yml)
    for project in $(yq e '.projects | keys | .[]' pancake.yml); do
        # Get the last updated date-time of the project
        project_folder="$project_location/$project"
        if [ -d "$project_folder" ]; then
            last_updated=$(git -C "$project_folder" log -1 --format="%cd")
            current_branch=$(git -C "$project_folder" rev-parse --abbrev-ref HEAD)
            last_committer=$(git -C "$project_folder" log -1 --format='%an')
            version=$(git -C "$project_folder" describe --tags --always)
            printf "| %-10s | %-10s | %-20s | %-10s | %-30s |\n" "$project" "$current_branch" "$last_committer" "$version" "$last_updated"
        else
            printf "| %-10s | %-10s | %-20s | %-10s | %-30s |\n" "$project" "-" "-" "-" "-"
        fi
    done
}

project_sync() {
    echo "🔄 Syncing projects..."
    # Parse pancake.yml and clone/update each project
    project_location=$(yq e '.project_location' pancake.yml)
    mkdir -p $project_location
    for project in $(yq e '.projects | keys | .[]' pancake.yml); do
        echo "🔄 Syncing $project..."
        project_folder="$project_location/$project"
        mkdir -p $project_folder
        git -C "$project_folder" pull || git clone "$(yq e ".projects.$project.github_link" pancake.yml)" "$project_folder"
    done
    echo "✅ All projects synced successfully."
}

project_sync_single() {
    project=$1
    echo "🔄 Syncing $project..."
    # Parse pancake.yml and clone/update the project
    project_location=$(yq e '.project_location' pancake.yml)
    project_folder="$project_location/$project"
    git -C "$project_folder" pull || git clone "$(yq e ".projects.$project.github_link" pancake.yml)" "$project_folder"
    echo "✅ $project synced successfully."
}

build_project() {
    project=$1
    echo "🔨 Building $project..."
    # Parse pancake.yml and get the build command for the project
    project_location=$(yq e '.project_location' pancake.yml)
    project_folder="$project_location/$project"
    build_command=$(yq e ".projects.$project.build" pancake.yml)
    if [ "$build_command" != "null" ]; then
        if [ -d "$project_folder" ]; then
            echo "Running in subshell: cd $project_folder && $build_command"
            (cd "$project_folder" && $build_command)
            echo "✅ $project built successfully."
        else
            echo "❌ The project directory does not exist."
        fi
    else
        echo "❌ Build variable not exists. Cannot build the project."
    fi
}

run_project() {
    project=$1
    echo "🏃 Running $project..."
    # Parse pancake.yml and get the run command for the project
    run_command=$(yq e ".projects.$project.run" pancake.yml)
    if [ "$run_command" != "null" ]; then
        # Replace all occurrences of @@variable@ with the value of the variable
        for variable in $(yq e 'keys | .[]' pancake.yml); do
            value=$(yq e ".$variable" pancake.yml)
            run_command=${run_command//@$variable@/$value}
        done
        # Replace <project_name> with the actual project name
        run_command=${run_command//<project_name>/$project}
        echo "Running: $run_command"
        $run_command
        echo "✅ $project run successfully."
    else
        echo "❌ Run variable not exists. Cannot run the project."
    fi
}

edit_config() {
    SUCCESS_MSG="✅ pancake.yml opened successfully."
    FAIL_MSG="❌ Failed to open pancake.yml."
    UNSUPPORTED_OS_MSG="❌ This OS is not supported."

    echo "🔧 Opening pancake.yml in the default editor..."
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        xdg-open pancake.yml && echo $SUCCESS_MSG || echo $FAIL_MSG
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        # Mac OSX
        open pancake.yml && echo $SUCCESS_MSG || echo $FAIL_MSG
    else
        echo $UNSUPPORTED_OS_MSG
        exit 1
    fi
}

help_menu() {
    echo "📖 Pancake Help Menu 📖"
    echo "Here are the available commands:"
    echo "  pancake status - Display the status"
    echo "  pancake project list - List all projects"
    echo "  pancake edit config - Edit the configuration"
    echo "  pancake run <project_name> - Run a specific project"
}

# Check the number of arguments and switch between different functions
if [ "$#" -eq 0 ]; then
    help_menu
    exit 0
elif [ "$1" = "status" ]; then
    status
elif [ "$1" = "project" ]; then
    if [ "$2" = "list" ]; then
        project_list
    elif [ "$2" = "sync" ]; then
        if [ -n "$3" ]; then
            project_sync_single $3
        else
            project_sync
        fi
    else
        echo "❌ Invalid second argument for project: $2"
        exit 1
    fi
elif [ "$1" = "edit" ]; then
    if [ "$2" = "config" ]; then
        edit_config
    else
        echo "❌ Invalid second argument for edit: $2"
        exit 1
    fi
elif [ "$1" = "run" ]; then
    if [ -n "$2" ]; then
        run_project $2
    else
        echo "⚠️ No second argument provided for run"
        exit 1
    fi
elif [ "$1" = "build" ]; then
    if [ -n "$2" ]; then
        build_project $2
    else
        echo "⚠️ No second argument provided for run"
        exit 1
    fi
else
    echo "❌ Invalid command: $1"
    exit 1
fi