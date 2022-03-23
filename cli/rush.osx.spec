# -*- mode: python ; coding: utf-8 -*-

block_cipher = None
hiddenimports = ["sentry_sdk.integrations.stdlib",
                 "sentry_sdk.integrations.excepthook",
                 "sentry_sdk.integrations.dedupe",
                 "sentry_sdk.integrations.atexit",
                 "sentry_sdk.integrations.modules",
                 "sentry_sdk.integrations.argv",
                 "sentry_sdk.integrations.logging",
                 "sentry_sdk.integrations.threading",
                 "sentry_sdk.integrations.django",
                 "sentry_sdk.integrations.flask",
                 "sentry_sdk.integrations.bottle",
                 "sentry_sdk.integrations.falcon",
                 "sentry_sdk.integrations.sanic",
                 "sentry_sdk.integrations.celery",
                 "sentry_sdk.integrations.rq",
                 "sentry_sdk.integrations.aiohttp",
                 "sentry_sdk.integrations.tornado",
                 "sentry_sdk.integrations.sqlalchemy",
                 "sentry_sdk.integrations.boto3"]



a = Analysis(['rush.py'],
             pathex=['.'],
             binaries=[],
             datas=[('icon.txt', '.'), ('zyzz.txt', '.'), ('external/deno_osx_64', 'external/'), ('external/zoneinfo.zip', 'external/'), ('adyen', 'adyen/'), ('adyen2', 'adyen2/'), ('wmt', 'wmt/'), ('paypal', 'paypal/'), ('datadome', 'datadome/'), ('rcfarm/mitm', 'mitm'),
              # ('bin/rcpool', 'bin/'), ('bin/rclogin', 'bin/'),
              ('bin/task_consumer', 'bin/'), ('external/ff_mac', 'firefox/'), ('is_harvest', '.')],
             hiddenimports=hiddenimports,
             hookspath=[],
             runtime_hooks=[],
             excludes=[],
             win_no_prefer_redirects=False,
             win_private_assemblies=False,
             cipher=block_cipher,
             noarchive=False)
pyz = PYZ(a.pure, a.zipped_data,
             cipher=block_cipher)
exe = EXE(pyz,
          a.scripts,
          a.binaries,
          a.zipfiles,
          a.datas,
          [],
          name='rush',
          debug=False,
          bootloader_ignore_signals=False,
          strip=False,
          upx=True,
          upx_exclude=[],
          runtime_tmpdir=None,
          console=True )
app = BUNDLE(exe,
             name='RushAIO.app',
             icon='rush.icns',
             bundle_identifier='com.rushaio.rushaio')
