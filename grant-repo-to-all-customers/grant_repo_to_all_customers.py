import os
import click
import github
import logging
from progress.bar import Bar

logging.basicConfig(format='%(asctime)s [%(levelname)s] %(message)s', level=logging.INFO)


@click.command()
@click.option(
    '--repo',
    prompt='Enter the name of the new repo in the gruntwork-io organization to grant teams access to',
    help='Name of the new repo in the gruntwork-io organization to grant teams access to.',
)
@click.option(
    '--seed',
    prompt='Enter the name of a repo in the Gruntwork organization to use as a reference for teams to grant access to',
    help='Name of a repo in the Gruntwork organization to use as a reference for teams to grant access to.',
)
@click.option(
    '--force/--prompt',
    default=False,
    help='When --force is passed in, skip the prompt to confirm if the team should be granted access.',
)
def main(repo, seed, force):
    """
    Grant access to a new repo to all customers who already have access to another repo in the library.
    """
    token = os.environ.get('GITHUB_OAUTH_TOKEN')
    if not token:
        raise click.ClickException('Environment variable GITHUB_OAUTH_TOKEN is not set')

    clt = github.GithubApiClient(token)

    logging.info(f'Retrieving all teams that have access to repo {seed}')
    teams = github.get_repository_teams(clt, seed)
    customer_teams = [t for t in teams if t.startswith('client-')]
    logging.info(f'Found {len(customer_teams)} customer teams that have access to {seed}')

    logging.info('The following teams will be granted access:')
    for team in customer_teams:
        logging.info(f'\t{team}')
    logging.info('')

    if force:
        logging.warn('--force flag was passed in. Skipping prompt.')
    else:
        input('Will grant access to the teams. [Ctrl+C] to cancel, or [ENTER] to proceed.')

    bar = Bar('Granting access', max=len(customer_teams))
    for team in bar.iter(customer_teams):
        github.add_repo_to_team(clt, team, repo)


if __name__ == '__main__':
    main()
